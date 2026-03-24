#!/usr/bin/env python3
"""
import_dsr.py — Imports all DSR items from DSR_Analizer_Ver.18.xls into xtmator.db

Usage:
    python3 cmd/import-dsr/import_dsr.py [--excel DSR_Analizer_Ver.18.xls] [--db xtmator.db]

Requirements:
    pip install xlrd   (xlrd 1.x/2.x — reads legacy .xls BIFF format)
    sqlite3 is part of Python standard library
"""

import sqlite3
import sys
import os
import argparse
import xlrd

# ── Chapter → Category name mapping ─────────────────────────────────────────
CHAPTER_CATEGORIES = {
    "2.":  "Earthwork",
    "4.":  "PCC",
    "5.":  "RCC",
    "6.":  "Brickwork",
    "7.":  "Stone Masonry",
    "8.":  "Marble & Tile Work",
    "9.":  "Woodwork & Joinery",
    "10.": "Steel Work",
    "11.": "Flooring",
    "12.": "Roofing",
    "13.": "Plastering",
    "14.": "Patch Plastering & Repairs",
    "15.": "Demolition",
    "16.": "Road Work",
    "17.": "Sanitary Fittings",
    "18.": "Plumbing (Pipes)",
    "19.": "Drainage",
    "20.": "Piling",
    "21.": "Aluminium Work",
    "22.": "Waterproofing",
    "23.": "Boring & Drilling",
    "24.": "Stone Surface Prep",
    "25.": "Aluminium Supply",
    "26.": "Bamboo / Timber Work",
    "30.": "Trenching",
    "50.": "Pumping & Dewatering",
    "51.": "Filling",
    "60.": "Ring Bund",
    "65.": "Equipment Hire",
}

# Canonical unit normalisation
UNIT_MAP = {
    "cum": "CUM", "m3": "CUM",
    "sqm": "SQM", "sqm.": "SQM", "m2": "SQM",
    "rmt": "M", "m": "M",
    "kg": "KG",
    "mt": "MT",
    "no.": "NO.", "no": "NO.",
    "l.s.": "LS", "ls": "LS",
    "day": "DAY", "hr": "HR", "month": "MONTH",
}
ALL_UNITS = set(UNIT_MAP.keys())


def normalise_unit(s: str) -> str:
    return UNIT_MAP.get(s.strip().lower(), s.strip().upper() or "CUM")


def is_unit(s: str) -> bool:
    return s.strip().lower() in ALL_UNITS


def is_numeric(s: str) -> bool:
    try:
        float(s.replace(",", ""))
        return True
    except ValueError:
        return False


def get_rate(row, ncols: int) -> float:
    """Prefer column 6 (total overhead rate), fallback to col 5 then 7."""
    for ci in [6, 5, 7]:
        if ci >= ncols:
            continue
        v = str(row[ci].value).strip().replace(",", "")
        try:
            r = float(v)
            if r > 0:
                return r
        except ValueError:
            pass
    return 0.0


def chapter_prefix(code: str) -> str:
    """Return e.g. '4.' from '4.1.2'"""
    dot: int = code.index(".")
    result: str = code[: dot + 1]  # type: ignore[index]
    return result


def parse_sheet(sh) -> list:
    """Parse the DAR Original sheet, return list of dicts."""
    items = []
    for r in range(sh.nrows):
        row = sh.row(r)
        if sh.ncols < 5:
            continue

        code = str(row[1].value).strip()
        if not code or "." not in code:
            continue

        # Only import known chapters
        try:
            ch: str = chapter_prefix(code)
            ch_num: float = float(ch[:-1])  # type: ignore[index]
        except (ValueError, IndexError):
            continue
        if ch not in CHAPTER_CATEGORIES:
            continue

        category = CHAPTER_CATEGORIES[ch]

        # Description — first non-numeric, non-unit text in cols 2..4
        desc = ""
        for ci in range(2, min(sh.ncols, 5)):
            v = str(row[ci].value).strip()
            if v and not is_numeric(v) and not is_unit(v):
                desc = " ".join(v.split())  # collapse whitespace
                break
        if not desc:
            continue

        # Unit — first unit-like string in cols 3..7
        unit = ""
        for ci in range(3, min(sh.ncols, 8)):
            v = str(row[ci].value).strip()
            if is_unit(v):
                unit = normalise_unit(v)
                break
        if not unit:
            unit = "CUM"

        rate = get_rate(row, sh.ncols)
        if rate <= 0:
            continue

        items.append({
            "category":    category,
            "code":        code,
            "description": desc,
            "unit":        unit,
            "rate":        rate,
        })

    return items


def import_to_db(items: list, db_path: str) -> int:
    conn = sqlite3.connect(db_path)
    cur = conn.cursor()

    # Ensure table exists (safe if already created by Go app)
    cur.execute("""
        CREATE TABLE IF NOT EXISTS dsr_items (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            category    TEXT NOT NULL,
            code        TEXT UNIQUE NOT NULL,
            description TEXT NOT NULL,
            unit        TEXT NOT NULL DEFAULT 'CUM',
            rate        REAL NOT NULL DEFAULT 0
        )
    """)

    n_imported: int = 0
    n_skipped: int = 0
    for item in items:
        try:
            cur.execute("""
                INSERT INTO dsr_items (category, code, description, unit, rate)
                VALUES (:category, :code, :description, :unit, :rate)
                ON CONFLICT(code) DO UPDATE SET
                    category    = excluded.category,
                    description = excluded.description,
                    unit        = excluded.unit,
                    rate        = excluded.rate
            """, item)
            n_imported += 1  # type: ignore[operator]
            print(f"  \u2705 [{item['code']}] {item['description'][:55]} \u2014 {item['unit']} @ \u20b9{item['rate']:.2f}")
        except Exception as e:
            print(f"  \u26a0  [{item['code']}] skipped: {e}")
            n_skipped += 1  # type: ignore[operator]

    conn.commit()
    conn.close()
    return n_imported


def main():
    parser = argparse.ArgumentParser(description="Import DSR items from .xls into SQLite")
    parser.add_argument("--excel", default="DSR_Analizer_Ver.18.xls", help="Path to DSR Excel file")
    parser.add_argument("--db",    default="xtmator.db",              help="Path to SQLite database")
    args = parser.parse_args()

    if not os.path.exists(args.excel):
        print(f"❌ Excel file not found: {args.excel}", file=sys.stderr)
        sys.exit(1)
    if not os.path.exists(args.db):
        print(f"❌ Database not found: {args.db}  (start the server once first to create it)", file=sys.stderr)
        sys.exit(1)

    print(f"📂 Opening Excel: {args.excel}")
    wb = xlrd.open_workbook(args.excel)

    # Find DAR Original sheet
    sh = None
    for s in wb.sheets():
        if "DAR" in s.name.upper() and "ORIGINAL" in s.name.upper():
            sh = s
            break
    if sh is None:
        sh = wb.sheets()[0]
    assert sh is not None, "No sheets found in workbook"
    print(f"\u2705 Using sheet: '{sh.name}' ({sh.nrows} rows)")

    print("\n🔍 Parsing items …")
    items = parse_sheet(sh)
    print(f"   Found {len(items)} valid items across {len(set(i['category'] for i in items))} categories\n")

    print(f"💾 Importing into: {args.db}")
    count = import_to_db(items, args.db)

    print(f"\n🎉 Done — {count} items imported/updated")

    # Summary by category
    from collections import Counter
    cats = Counter(i["category"] for i in items)
    print("\n📊 Category summary:")
    for cat, n in sorted(cats.items()):
        print(f"   {cat:<35} {n:>4} items")


if __name__ == "__main__":
    main()
