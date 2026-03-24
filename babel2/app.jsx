    const { useState, useEffect, useCallback, useRef } = React;

    // ─── API helpers ────────────────────────────────────────────────────────────
    const getToken = () => localStorage.getItem('xtm_token');
    const getUser  = () => { try { return JSON.parse(localStorage.getItem('xtm_user')||'null'); } catch { return null; } };

    async function api(method, path, body) {
      const headers = { 'Content-Type': 'application/json' };
      const token = getToken();
      if (token) headers['Authorization'] = 'Bearer ' + token;
      const res = await fetch(path, { method, headers, body: body ? JSON.stringify(body) : undefined });
      const json = await res.json();
      if (!res.ok) throw new Error(json.error || 'An error occurred');
      return json;
    }

    // ─── Toast ───────────────────────────────────────────────────────────────────
    function Toast({ msg, type, onDone }) {
      useEffect(() => { const t = setTimeout(onDone, 3000); return () => clearTimeout(t); }, []);
      return <div className={'toast toast-' + type}>{msg}</div>;
    }

    function useToast() {
      const [toast, setToast] = useState(null);
      const show = (msg, type='success') => setToast({ msg, type, key: Date.now() });
      const el = toast ? <Toast key={toast.key} msg={toast.msg} type={toast.type} onDone={() => setToast(null)}/> : null;
      return [el, show];
    }

    // ─── Auth Pages ───────────────────────────────────────────────────────────────
    function AuthPage({ onLogin }) {
      const [mode, setMode] = useState('login'); // 'login' | 'register'
      const [form, setForm] = useState({ name:'', email:'', password:'' });
      const [loading, setLoading] = useState(false);
      const [error, setError] = useState('');

      // Password strength for registration UX
      const pwStrength = React.useMemo(() => {
        const p = form.password;
        if (!p) return null;
        const checks = [
          { label: '8+ characters', ok: p.length >= 8 },
          { label: 'Uppercase letter', ok: /[A-Z]/.test(p) },
          { label: 'Lowercase letter', ok: /[a-z]/.test(p) },
          { label: 'Number', ok: /[0-9]/.test(p) },
        ];
        return checks;
      }, [form.password]);

      const set = (k) => (e) => setForm(f => ({ ...f, [k]: e.target.value }));

      const submit = async (e) => {
        e.preventDefault();
        setLoading(true); setError('');
        try {
          if (mode === 'register') {
            await api('POST', '/api/auth/register', form);
            setMode('login');
            setForm(f => ({ ...f, name: '' }));
          } else {
            const res = await api('POST', '/api/auth/login', { email: form.email, password: form.password });
            localStorage.setItem('xtm_token', res.token);
            localStorage.setItem('xtm_user', JSON.stringify(res.user));
            onLogin(res.user);
          }
        } catch(err) {
          setError(err.message);
        } finally { setLoading(false); }
      };

      return (
        <div style={{ minHeight:'100vh', display:'flex', alignItems:'center', justifyContent:'center', padding:20 }}>
          <div style={{ width:'100%', maxWidth:420 }}>
            {/* Logo */}
            <div style={{ textAlign:'center', marginBottom:32 }}>
              <h1 style={{ fontSize:32, fontWeight:800, background:'linear-gradient(135deg,#3b82f6,#60a5fa)', WebkitBackgroundClip:'text', WebkitTextFillColor:'transparent' }}>
                XTMATOR
              </h1>
              <p style={{ color:'var(--text-muted)', fontSize:14, marginTop:6 }}>
                Construction BOQ Estimation System
              </p>
            </div>
            <div className="card">
              {/* Tab switcher */}
              <div style={{ display:'flex', borderRadius:8, background:'var(--surface2)', padding:3, marginBottom:24 }}>
                {['login','register'].map(m => (
                  <button key={m} onClick={() => { setMode(m); setError(''); }}
                    style={{ flex:1, padding:'8px 0', border:'none', borderRadius:6, cursor:'pointer', fontWeight:600, fontSize:14, fontFamily:'inherit',
                      background: mode === m ? 'var(--accent)' : 'transparent',
                      color: mode === m ? '#fff' : 'var(--text-muted)', transition:'all .15s' }}>
                    {m === 'login' ? 'Sign In' : 'Create Account'}
                  </button>
                ))}
              </div>

              {/* Google Sign-In button */}
              <a href="/api/auth/google"
                id="btn-google"
                style={{
                  display:'flex', alignItems:'center', justifyContent:'center', gap:10,
                  padding:'10px 16px', borderRadius:8, border:'1px solid var(--border)',
                  background:'#fff', color:'#3c4043', fontWeight:600, fontSize:14,
                  textDecoration:'none', marginBottom:16, transition:'box-shadow .15s',
                  cursor:'pointer',
                }}
                onMouseEnter={e => e.currentTarget.style.boxShadow='0 2px 8px rgba(0,0,0,.2)'}
                onMouseLeave={e => e.currentTarget.style.boxShadow='none'}
              >
                {/* Google G logo SVG */}
                <svg width="18" height="18" viewBox="0 0 48 48">
                  <path fill="#EA4335" d="M24 9.5c3.5 0 6.6 1.2 9.1 3.2l6.8-6.8C35.8 2.2 30.2 0 24 0 14.7 0 6.7 5.4 2.8 13.3l7.9 6.1C12.6 13 17.9 9.5 24 9.5z"/>
                  <path fill="#4285F4" d="M46.5 24.5c0-1.6-.1-3.1-.4-4.5H24v8.5h12.7c-.6 3-2.3 5.5-4.8 7.2l7.6 5.9C43.7 37.5 46.5 31.5 46.5 24.5z"/>
                  <path fill="#FBBC05" d="M10.7 28.6A14.4 14.4 0 0 1 9.5 24c0-1.6.3-3.2.8-4.6L2.4 13.3A23.9 23.9 0 0 0 0 24c0 3.9.9 7.6 2.8 10.8l7.9-6.2z"/>
                  <path fill="#34A853" d="M24 48c6.2 0 11.4-2 15.2-5.5l-7.6-5.9c-2 1.4-4.6 2.2-7.6 2.2-6.1 0-11.3-4.1-13.2-9.7l-7.9 6.2C6.7 42.6 14.7 48 24 48z"/>
                </svg>
                Continue with Google
              </a>

              {/* Divider */}
              <div style={{ display:'flex', alignItems:'center', gap:10, marginBottom:16 }}>
                <div style={{ flex:1, height:1, background:'var(--border)' }}/>
                <span style={{ color:'var(--text-dim)', fontSize:12 }}>or with email</span>
                <div style={{ flex:1, height:1, background:'var(--border)' }}/>
              </div>

              <form onSubmit={submit}>
                {mode === 'register' && (
                  <div className="form-group">
                    <label className="label">Full Name</label>
                    <input id="inp-name" className="input" placeholder="John Builder" value={form.name} onChange={set('name')} required/>
                  </div>
                )}
                <div className="form-group">
                  <label className="label">Email Address</label>
                  <input id="inp-email" className="input" type="email" placeholder="you@example.com" value={form.email} onChange={set('email')} required/>
                </div>
                <div className="form-group">
                  <label className="label">Password</label>
                  <input id="inp-password" className="input" type="password" placeholder="••••••••" value={form.password} onChange={set('password')} required/>
                  {/* Password strength indicator for registration */}
                  {mode === 'register' && pwStrength && (
                    <div style={{ display:'flex', gap:8, flexWrap:'wrap', marginTop:8 }}>
                      {pwStrength.map(c => (
                        <span key={c.label} style={{
                          fontSize:11, padding:'2px 8px', borderRadius:20,
                          background: c.ok ? 'rgba(34,197,94,.15)' : 'var(--surface2)',
                          color: c.ok ? 'var(--success)' : 'var(--text-dim)',
                          border: '1px solid ' + (c.ok ? 'rgba(34,197,94,.3)' : 'var(--border)'),
                          transition:'all .2s',
                        }}>
                          {c.ok ? '✓' : '○'} {c.label}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
                {error && <p style={{ color:'var(--danger)', fontSize:13, marginBottom:14 }}>⚠ {error}</p>}
                <button id="btn-submit" className="btn btn-primary" style={{ width:'100%', justifyContent:'center', padding:'11px 0', fontSize:15 }} disabled={loading}>
                  {loading ? <span className="spinner"/> : (mode === 'login' ? 'Sign In' : 'Create Account')}
                </button>
              </form>
            </div>
          </div>
        </div>
      );
    }

    // ─── Sidebar ─────────────────────────────────────────────────────────────────
    function Sidebar({ user, page, onNav, onLogout, projects, currentProject, onOpenProject }) {
      const initials = (user?.name||'U').split(' ').map(w=>w[0]).join('').toUpperCase().slice(0,2);
      const recent = (projects || []).slice(0, 5);

      return (
        <div className="sidebar">
          <div className="sidebar-logo">
            <h1>XTMATOR</h1>
            <p>BOQ Estimation System</p>
          </div>
          <div className="sidebar-nav">
            <button id="nav-dashboard" className={'nav-item' + (page==='dashboard'?' active':'')} onClick={() => onNav('dashboard')}>
              🗂 Dashboard
            </button>
            {recent.length > 0 && (
              <div style={{ marginTop: 24, paddingLeft: 12, marginBottom: 8, fontSize: 11, fontWeight: 700, color: 'var(--text-dim)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                Recent Projects
              </div>
            )}
            {recent.map(p => (
              <button key={p.ID} className={'nav-item' + (page==='project' && currentProject?.ID === p.ID ? ' active' : '')} style={{ padding: '8px 12px', fontSize: 13, gap: 8 }} onClick={() => onOpenProject(p)}>
                <span style={{ fontSize: 14 }}>📄</span>
                <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{p.Name}</span>
              </button>
            ))}
          </div>
          <div className="sidebar-user">
            <div className="user-info">
              <div className="avatar">{initials}</div>
              <div style={{ flex:1, minWidth:0 }}>
                <div style={{ fontWeight:600, fontSize:13, overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{user?.name}</div>
                <div style={{ fontSize:11, color:'var(--text-muted)', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{user?.email}</div>
              </div>
            </div>
            <button id="btn-logout" className="btn btn-outline btn-sm" style={{ width:'100%', justifyContent:'center', marginTop:12 }} onClick={onLogout}>
              Sign Out
            </button>
          </div>
        </div>
      );
    }

    // ─── Dashboard page ───────────────────────────────────────────────────────────
    function Dashboard({ projects, loading, onRefresh, onOpenProject, showToast }) {
      const [showModal, setShowModal] = useState(false);
      const [form, setForm] = useState({ name:'', client_name:'', location:'' });
      const [saving, setSaving] = useState(false);

      const createProject = async (e) => {
        e.preventDefault(); setSaving(true);
        try {
          const p = await api('POST', '/api/projects', form);
          setShowModal(false);
          setForm({ name:'', client_name:'', location:'' });
          showToast('Project created!');
          onRefresh();
          onOpenProject(p);
        } catch(err) { showToast(err.message, 'error'); }
        finally { setSaving(false); }
      };

      const deleteProject = async (e, id) => {
        e.stopPropagation();
        if (!confirm('Delete this project and all its BOQ items?')) return;
        try { await api('DELETE', '/api/projects/' + id); onRefresh(); showToast('Project deleted'); }
        catch(err) { showToast(err.message, 'error'); }
      };

      const set = k => e => setForm(f => ({ ...f, [k]: e.target.value }));

      return (
        <div>
          <div className="page-header">
            <div>
              <h2>My Projects</h2>
              <p style={{ color:'var(--text-muted)', fontSize:14, marginTop:2 }}>Select a project to open its BOQ sheet</p>
            </div>
            <button id="btn-new-project" className="btn btn-primary" onClick={() => setShowModal(true)}>+ New Project</button>
          </div>

          {loading ? (
            <div style={{ textAlign:'center', padding:60 }}><span className="spinner"/></div>
          ) : projects.length === 0 ? (
            <div className="empty-state">
              <div className="icon">🏗</div>
              <h3>No projects yet</h3>
              <p>Create your first project to start building BOQ sheets</p>
              <button className="btn btn-primary" style={{ marginTop:16 }} onClick={() => setShowModal(true)}>+ Create Project</button>
            </div>
          ) : (
            <div className="project-grid">
              {projects.map(p => (
                <div key={p.ID} className="project-card" onClick={() => onOpenProject(p)}>
                  <div style={{ display:'flex', justifyContent:'space-between', alignItems:'start' }}>
                    <h3>{p.Name}</h3>
                    <button className="btn btn-danger btn-sm" onClick={e => deleteProject(e, p.ID)}>🗑</button>
                  </div>
                  {p.ClientName && <p style={{ marginTop:4 }}>👤 {p.ClientName}</p>}
                  {p.Location && <p style={{ marginTop:2 }}>📍 {p.Location}</p>}
                  <p style={{ marginTop:8, fontSize:11 }}>
                    {new Date(p.CreatedAt).toLocaleDateString('en-IN', { day:'numeric', month:'short', year:'numeric' })}
                  </p>
                </div>
              ))}
            </div>
          )}

          {showModal && (
            <div className="modal-overlay" onClick={e => e.target===e.currentTarget && setShowModal(false)}>
              <div className="modal">
                <div className="modal-header">
                  <h3>New Project</h3>
                  <button className="close-btn" onClick={() => setShowModal(false)}>×</button>
                </div>
                <form onSubmit={createProject}>
                  <div className="form-group">
                    <label className="label">Project Name *</label>
                    <input id="inp-project-name" className="input" placeholder="e.g. Water Tank – Block A" value={form.name} onChange={set('name')} required/>
                  </div>
                  <div className="form-group">
                    <label className="label">Client Name</label>
                    <input id="inp-client-name" className="input" placeholder="e.g. Kerala Water Authority" value={form.client_name} onChange={set('client_name')}/>
                  </div>
                  <div className="form-group">
                    <label className="label">Location / Site</label>
                    <input id="inp-location" className="input" placeholder="e.g. Thrissur, Kerala" value={form.location} onChange={set('location')}/>
                  </div>
                  <div style={{ display:'flex', gap:10, justifyContent:'flex-end', marginTop:8 }}>
                    <button type="button" className="btn btn-outline" onClick={() => setShowModal(false)}>Cancel</button>
                    <button id="btn-create-project" type="submit" className="btn btn-primary" disabled={saving}>
                      {saving ? <span className="spinner"/> : 'Create Project'}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          )}
        </div>
      );
    }

    // ─── BOQ Entry Form Modal ─────────────────────────────────────────────────────
    // dimMode: determines which dimension inputs to show based on unit
    function getDimMode(unit) {
      const u = (unit||'').toUpperCase().trim();
      if (u === 'CUM' || u === 'M3') return '3d';
      if (u === 'SQM' || u === 'SQM.' || u === 'M2') return '2d';
      if (u === 'M' || u === 'RMT') return '1d';
      return '0d'; // KG, MT, NO., LS, DAY, HR, MONTH etc.
    }

    // Human-readable hint for each dim mode
    function dimHint(unit, mode) {
      if (mode === '3d') return 'Enter L \u00d7 B \u00d7 H to auto-calculate Qty in ' + unit;
      if (mode === '2d') return 'Enter L \u00d7 B to auto-calculate Qty in ' + unit;
      if (mode === '1d') return 'Enter Length to auto-calculate Qty in ' + unit;
      return 'Enter Quantity directly in ' + unit;
    }


    function BOQEntryModal({ projectID, sheetID, onSaved, onClose, showToast }) {
      const [categories, setCategories] = useState([]);
      const [items, setItems] = useState([]);
      const [customItems, setCustomItems] = useState([]);
      const [saveToLibrary, setSaveToLibrary] = useState(false);
      const [form, setForm] = useState({
        dsr_item_id: '', category:'', description:'', length:'', breadth:'', height:'', manual_qty:'', manual_rate:'', manual_unit:'CUM', manual_category:''
      });
      const [preview, setPreview] = useState({ qty:0, amount:0 });
      const [saving, setSaving] = useState(false);

      useEffect(() => {
        api('GET', '/api/dsr/categories').then(setCategories).catch(() => {});
        api('GET', '/api/custom-items').then(setCustomItems).catch(() => {});
      }, []);

      useEffect(() => {
        if (!form.category) { setItems([]); return; }
        api('GET', '/api/dsr/items?category=' + encodeURIComponent(form.category))
          .then(setItems).catch(() => {});
      }, [form.category]);

      // Derive the active unit and dim-mode from selected DSR item (or manual)
      const selectedItem = items.find(i => String(i.ID) === form.dsr_item_id);
      const activeUnit = selectedItem ? selectedItem.Unit : (form.category === '__custom__' ? form.manual_unit : '');
      const dimMode = getDimMode(activeUnit);

      useEffect(() => {
        const l = parseFloat(form.length)||0, b = parseFloat(form.breadth)||0, h = parseFloat(form.height)||0;
        const mq = parseFloat(form.manual_qty)||0;
        let qty = 0;
        if (dimMode === '3d' && l > 0 && b > 0 && h > 0) qty = parseFloat((l*b*h).toFixed(4));
        else if (dimMode === '2d' && l > 0 && b > 0) qty = parseFloat((l*b).toFixed(4));
        else if (dimMode === '1d' && l > 0) qty = parseFloat(l.toFixed(4));
        if (qty === 0) qty = mq;
        const rate = parseFloat(form.manual_rate) || (selectedItem?.Rate||0);
        setPreview({ qty, amount: parseFloat((qty * rate).toFixed(2)) });
      }, [form, selectedItem, dimMode]);

      const set = k => e => {
        const v = e.target.value;
        const updated = { ...form, [k]: v };
        if (k === 'dsr_item_id') {
          const sel = items.find(i => String(i.ID) === v);
          if (sel) { updated.description = '[' + sel.Code + '] ' + sel.Description; updated.manual_rate = ''; }
        }
        if (k === 'category') { updated.dsr_item_id = ''; updated.description = ''; }
        setForm(updated);
      };

      const submit = async (e) => {
        e.preventDefault(); setSaving(true);
        const body = {
          sheet_id: sheetID,
          dsr_item_id: form.dsr_item_id ? parseInt(form.dsr_item_id) : null,
          description: form.description,
          category: form.category === '__custom__' ? (form.manual_category || 'Custom') : form.category,
          unit: form.category === '__custom__' ? form.manual_unit : '',
          length: parseFloat(form.length)||0,
          breadth: parseFloat(form.breadth)||0,
          height: parseFloat(form.height)||0,
          manual_qty: parseFloat(form.manual_qty)||0,
          manual_rate: parseFloat(form.manual_rate)||0,
        };
        try {
          if (form.category === '__custom__' && saveToLibrary) {
            await api('POST', '/api/custom-items', {
              category: form.manual_category || 'Custom',
              description: form.description,
              unit: form.manual_unit,
              rate: parseFloat(form.manual_rate) || 0
            });
          }
          await api('POST', '/api/projects/' + projectID + '/boq', body);
          showToast('Item added!'); onSaved();
        } catch(err) { showToast(err.message, 'error'); }
        finally { setSaving(false); }
      };

      const unitLabel = activeUnit || '—';

      return (
        <div className="modal-overlay" onClick={e => e.target===e.currentTarget && onClose()}>
          <div className="modal" style={{ maxWidth:640 }}>
            <div className="modal-header">
              <h3>Add BOQ Item</h3>
              <button className="close-btn" onClick={onClose}>×</button>
            </div>
            <form onSubmit={submit}>
              {/* Category + DSR item */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 16 }}>
                <div className="form-group" style={{ marginBottom: 0 }}>
                  <label className="label">Category *</label>
                  <select id="sel-category" className="select" value={form.category} onChange={set('category')} required>
                    <option value="">— Select category —</option>
                    {categories.map(c => <option key={c} value={c}>{c}</option>)}
                    <option value="__custom__">Custom / Manual</option>
                  </select>
                </div>
                <div className="form-group" style={{ marginBottom: 0 }}>
                  <label className="label">DSR Work Item (Check DSOR)</label>
                  <select id="sel-dsr-item" className="select" value={form.dsr_item_id} onChange={set('dsr_item_id')} disabled={!form.category || form.category==='__custom__'}>
                    <option value="">— Select work item —</option>
                    {items.map(i => <option key={i.ID} value={i.ID}>[{i.Code}] {i.Description} — ₹{i.Rate.toLocaleString('en-IN')}/{i.Unit}</option>)}
                  </select>
                </div>
              </div>

              {form.category === '__custom__' && customItems.length > 0 && (
                <div className="form-group" style={{ marginBottom: 12 }}>
                  <label className="label">My Custom Library</label>
                  <select className="select" onChange={e => {
                      const sel = customItems.find(c => String(c.id) === e.target.value);
                      if (sel) {
                        setForm(f => ({...f, 
                          description: sel.description, 
                          manual_category: sel.category,
                          manual_rate: sel.rate,
                          manual_unit: sel.unit || 'CUM'
                        }));
                      }
                  }}>
                    <option value="">— Load saved item —</option>
                    {customItems.map(c => <option key={c.id} value={c.id}>[{c.category}] {c.description} - ₹{c.rate}/{c.unit}</option>)}
                  </select>
                </div>
              )}

              {/* Custom category name when manual */}
              {form.category === '__custom__' && (
                <div style={{ display:'grid', gridTemplateColumns:'2fr 1fr', gap:12, marginBottom:16 }}>
                  <div className="form-group" style={{ marginBottom:0 }}>
                    <label className="label">Custom Category Name</label>
                    <input className="input" placeholder="e.g. Miscellaneous" value={form.manual_category||''} onChange={e => setForm(f => ({...f, manual_category: e.target.value}))}/>
                  </div>
                  <div className="form-group" style={{ marginBottom:0 }}>
                    <label className="label">Unit</label>
                    <select className="select" value={form.manual_unit} onChange={set('manual_unit')}>
                      <option value="CUM">CUM / M3</option>
                      <option value="SQM">SQM / M2</option>
                      <option value="M">M / RMT</option>
                      <option value="KG">KG</option>
                      <option value="MT">MT</option>
                      <option value="NO.">NO.</option>
                      <option value="LS">LS</option>
                      <option value="DAY">DAY</option>
                      <option value="MONTH">MONTH</option>
                    </select>
                  </div>
                </div>
              )}

              {/* Description */}
              <div className="form-group">
                <label className="label">Description *</label>
                <input id="inp-desc" className="input" placeholder="Description of work item" value={form.description} onChange={set('description')} required/>
              </div>

              {/* Selected item info bar */}
              {selectedItem && (
                <div style={{ background:'var(--surface2)', borderRadius:8, padding:'10px 14px', marginBottom:16, fontSize:13, display:'flex', gap:16, alignItems:'center', flexWrap:'wrap' }}>
                  <span><span className="tag">{selectedItem.Code}</span></span>
                  <span style={{ color:'var(--text-muted)' }}>DSR Rate:</span>
                  <span style={{ color:'var(--accent)', fontWeight:600 }}>₹{selectedItem.Rate.toLocaleString('en-IN')} / {selectedItem.Unit}</span>
                  <span className="badge badge-blue">{selectedItem.Unit}</span>
                </div>
              )}

              <div className="divider"/>

              {/* Hint text — context-sensitive */}
              {activeUnit && (
                <p style={{ fontSize:12, color:'var(--text-muted)', marginBottom:12 }}>
                  {dimHint(unitLabel, dimMode)}
                </p>
              )}

              {/* ── Dimension fields (shown based on unit) ── */}
              {dimMode === '3d' && (
                <div className="boq-form-grid" style={{ marginBottom:12 }}>
                  <div className="form-group" style={{ marginBottom:0 }}>
                    <label className="label">Length (m)</label>
                    <input id="inp-length" className="input" type="number" step="0.001" placeholder="0.000" value={form.length} onChange={set('length')}/>
                  </div>
                  <div className="form-group" style={{ marginBottom:0 }}>
                    <label className="label">Breadth (m)</label>
                    <input id="inp-breadth" className="input" type="number" step="0.001" placeholder="0.000" value={form.breadth} onChange={set('breadth')}/>
                  </div>
                  <div className="form-group" style={{ marginBottom:0 }}>
                    <label className="label">Height / Depth (m)</label>
                    <input id="inp-height" className="input" type="number" step="0.001" placeholder="0.000" value={form.height} onChange={set('height')}/>
                  </div>
                </div>
              )}

              {dimMode === '2d' && (
                <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr 1fr', gap:12, marginBottom:12 }}>
                  <div className="form-group" style={{ marginBottom:0 }}>
                    <label className="label">Length (m)</label>
                    <input id="inp-length" className="input" type="number" step="0.001" placeholder="0.000" value={form.length} onChange={set('length')}/>
                  </div>
                  <div className="form-group" style={{ marginBottom:0 }}>
                    <label className="label">Breadth (m)</label>
                    <input id="inp-breadth" className="input" type="number" step="0.001" placeholder="0.000" value={form.breadth} onChange={set('breadth')}/>
                  </div>
                  <div style={{ display:'flex', alignItems:'center', justifyContent:'center', color:'var(--text-dim)', fontSize:12, paddingTop:20 }}>
                    Height not needed for {unitLabel}
                  </div>
                </div>
              )}

              {dimMode === '1d' && (
                <div style={{ display:'grid', gridTemplateColumns:'1fr 2fr', gap:12, marginBottom:12 }}>
                  <div className="form-group" style={{ marginBottom:0 }}>
                    <label className="label">Length (m)</label>
                    <input id="inp-length" className="input" type="number" step="0.001" placeholder="0.000" value={form.length} onChange={set('length')}/>
                  </div>
                  <div style={{ display:'flex', alignItems:'center', color:'var(--text-dim)', fontSize:12, paddingTop:20 }}>
                    Only length needed for {unitLabel} items
                  </div>
                </div>
              )}

              {/* Manual qty + rate override */}
              <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:12, marginBottom:16 }}>
                <div className="form-group" style={{ marginBottom:0 }}>
                  <label className="label">
                    {dimMode === '0d' ? 'Quantity (' + (unitLabel || 'units') + ') *' : 'Manual Qty (' + (unitLabel || 'units') + ')'}
                  </label>
                  <input
                    id="inp-manual-qty"
                    className="input"
                    type="number"
                    step="0.001"
                    placeholder={dimMode === '0d' ? 'Required' : 'Override if not using dimensions'}
                    value={form.manual_qty}
                    onChange={set('manual_qty')}
                    required={dimMode === '0d'}
                  />
                </div>
                <div className="form-group" style={{ marginBottom:0 }}>
                  <label className="label">Rate Override (₹/{unitLabel || 'unit'})</label>
                  <input
                    id="inp-manual-rate"
                    className="input"
                    type="number"
                    step="0.01"
                    placeholder={selectedItem ? String(selectedItem.Rate) : 'Enter rate'}
                    value={form.manual_rate}
                    onChange={set('manual_rate')}
                    required={!selectedItem}
                  />
                </div>
              </div>

              {/* Live preview */}
              {(preview.qty > 0) && (
                <div style={{ background:'var(--accent-soft)', border:'1px solid var(--accent)', borderRadius:8, padding:'12px 16px', marginBottom:16, display:'flex', justifyContent:'space-between', fontSize:13 }}>
                  <span>Qty: <strong>{preview.qty.toFixed(3)} {unitLabel}</strong></span>
                  <span>Amount: <strong style={{ color:'var(--accent)' }}>₹{preview.amount.toLocaleString('en-IN', { minimumFractionDigits:2 })}</strong></span>
                </div>
              )}

              {form.category === '__custom__' && (
                <div style={{ marginBottom:16, display:'flex', alignItems:'center', gap:8 }}>
                  <input type="checkbox" id="chk-save-lib" checked={saveToLibrary} onChange={e => setSaveToLibrary(e.target.checked)} />
                  <label htmlFor="chk-save-lib" style={{ fontSize:13, cursor:'pointer' }}>Save this custom item to my library for future use</label>
                </div>
              )}
              <div style={{ display:'flex', gap:10, justifyContent:'flex-end' }}>
                <button type="button" className="btn btn-outline" onClick={onClose}>Cancel</button>
                <button id="btn-add-item" type="submit" className="btn btn-primary" disabled={saving}>
                  {saving ? <span className="spinner"/> : 'Add to BOQ'}
                </button>
              </div>
            </form>
          </div>
        </div>
      );
    }

    // ─── Project Sheet (BOQ) ──────────────────────────────────────────────────────
    function ProjectSheet({ project, onBack, showToast }) {
      const [sheet, setSheet] = useState(null);
      const [loading, setLoading] = useState(true);
      const [sheets, setSheets] = useState([]);
      const [activeSheetID, setActiveSheetID] = useState(0);
      const [loadingSheets, setLoadingSheets] = useState(true);
      const [showAdd, setShowAdd] = useState(false);
      const [exporting, setExporting] = useState(false);
      const [costIndex, setCostIndex] = useState(project.CostIndex || 0);
      const [editingRow, setEditingRow] = useState(null);
      const [editForm, setEditForm] = useState({ length:'', breadth:'', height:'', manual_qty:'', manual_rate:'' });
      const [savingEdit, setSavingEdit] = useState(false);
      const [savingIndex, setSavingIndex] = useState(false);

      const loadSheets = useCallback(async () => {
        setLoadingSheets(true);
        try {
          const list = await api('GET', '/api/projects/' + project.ID + '/sheets');
          setSheets(list);
          if (list.length > 0 && activeSheetID === 0) {
            setActiveSheetID(list[0].id);
          }
        } catch(e) { showToast(e.message, 'error'); }
        finally { setLoadingSheets(false); }
      }, [project.ID, activeSheetID]);

      useEffect(() => { loadSheets(); }, [loadSheets]);

      const load = useCallback(async () => {
        if (!activeSheetID) return;
        setLoading(true);
        try { setSheet(await api('GET', '/api/projects/' + project.ID + '/boq?sheet_id=' + activeSheetID)); }
        catch(e) { showToast(e.message, 'error'); }
        finally { setLoading(false); }
      }, [project.ID, activeSheetID]);

      useEffect(() => { load(); }, [load]);

      const deleteEntry = async (id) => {
        if (!confirm('Remove this item?')) return;
        try { await api('DELETE', '/api/projects/' + project.ID + '/boq/' + id); load(); showToast('Item removed'); }
        catch(e) { showToast(e.message, 'error'); }
      };

      const updateCostIndex = async () => {
        setSavingIndex(true);
        try {
          await api('PUT', '/api/projects/' + project.ID, {
            name: project.Name,
            client_name: project.ClientName,
            location: project.Location,
            cost_index: parseFloat(costIndex) || 0
          });
          project.CostIndex = parseFloat(costIndex) || 0;
          showToast('Cost Index applied!');
          load();
        } catch(e) { showToast(e.message, 'error'); }
        finally { setSavingIndex(false); }
      };

      const exportPDF = async () => {
        setExporting(true);
        try {
          const token = getToken();
          const res = await fetch('/api/projects/' + project.ID + '/export/pdf?sheet_id=' + activeSheetID, {
            headers: { 'Authorization': 'Bearer ' + token }
          });
          if (!res.ok) { const j = await res.json(); throw new Error(j.error); }
          const blob = await res.blob();
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a'); a.href = url;
          a.download = 'BOQ_' + project.Name.replace(/\s+/g,'_') + '.pdf';
          a.click(); URL.revokeObjectURL(url);
          showToast('PDF downloaded!');
        } catch(e) { showToast(e.message, 'error'); }
        finally { setExporting(false); }
      };

      const saveEdit = async (id) => {
        setSavingEdit(true);
        try {
          await api('PUT', '/api/projects/' + project.ID + '/boq/' + id, {
            length: parseFloat(editForm.length) || 0,
            breadth: parseFloat(editForm.breadth) || 0,
            height: parseFloat(editForm.height) || 0,
            manual_qty: parseFloat(editForm.manual_qty) || 0,
            manual_rate: parseFloat(editForm.manual_rate) || 0
          });
          setEditingRow(null);
          load();
          showToast('Item updated!');
        } catch(e) { showToast(e.message, 'error'); }
        finally { setSavingEdit(false); }
      };

      const startEdit = (item) => {
        setEditingRow(item.ID);
        const ratio = 1 + ((project.CostIndex || 0) / 100.0);
        const rawRate = item.Rate / ratio;
        setEditForm({
          length: item.Length > 0 ? item.Length : '',
          breadth: item.Breadth > 0 ? item.Breadth : '',
          height: item.Height > 0 ? item.Height : '',
          manual_qty: item.Quantity,
          manual_rate: rawRate.toFixed(2)
        });
      };

      const setEdit = k => e => setEditForm(f => ({ ...f, [k]: e.target.value }));

      const exportExcel = async () => {
        setExporting(true);
        try {
          const token = getToken();
          const res = await fetch('/api/projects/' + project.ID + '/export/excel?sheet_id=' + activeSheetID, {
            headers: { 'Authorization': 'Bearer ' + token }
          });
          if (!res.ok) { const j = await res.json(); throw new Error(j.error); }
          const blob = await res.blob();
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a'); a.href = url;
          a.download = 'BOQ_' + project.Name.replace(/\s+/g,'_') + '.xlsx';
          a.click(); URL.revokeObjectURL(url);
          showToast('Excel downloaded!');
        } catch(e) { showToast(e.message, 'error'); }
        finally { setExporting(false); }
      };

      const fmt = (n) => (n||0).toLocaleString('en-IN', { minimumFractionDigits:2, maximumFractionDigits:2 });

      return (
        <div>
          <div className="page-header">
            <div style={{ display:'flex', alignItems:'center', gap:12 }}>
              <button id="btn-back" className="btn btn-outline btn-sm" onClick={onBack}>← Back</button>
              <div>
                <div style={{ display:'flex', alignItems:'center', gap: 16 }}>
                  <h2>{project.Name}</h2>
                  <div style={{ display:'flex', alignItems:'center', background:'var(--surface2)', padding:'4px 8px', borderRadius:8, border:'1px solid var(--border)' }}>
                    <span style={{ fontSize:12, fontWeight:600, color:'var(--text-muted)', marginRight:6 }}>INDEX:</span>
                    <input 
                      type="number" 
                      style={{ width:60, background:'transparent', border:'none', color:'var(--accent)', fontWeight:700, outline:'none' }}
                      value={costIndex}
                      onChange={e => setCostIndex(e.target.value)}
                      onBlur={() => {
                        if (parseFloat(costIndex) !== (project.CostIndex || 0)) updateCostIndex();
                      }}
                      onKeyDown={e => {
                        if (e.key === 'Enter' && parseFloat(costIndex) !== (project.CostIndex || 0)) updateCostIndex();
                      }}
                    />
                    <span style={{ fontSize:14, fontWeight:700, color:'var(--text-muted)' }}>%</span>
                    {savingIndex && <span className="spinner" style={{ width:12, height:12, marginLeft:8 }}/>}
                  </div>
                </div>
                <p style={{ color:'var(--text-muted)', fontSize:13, marginTop:2 }}>
                  {project.ClientName && <span>👤 {project.ClientName}  </span>}
                  {project.Location && <span>📍 {project.Location}</span>}
                </p>
              </div>
            </div>
            <div style={{ display:'flex', gap:10 }}>
              <button id="btn-share" className="btn btn-outline" onClick={async () => {
                try {
                  const res = await api('POST', '/api/projects/' + project.ID + '/share');
                  const url = window.location.origin + '/share/' + res.share_token;
                  navigator.clipboard.writeText(url);
                  showToast('🔗 Share link copied to clipboard!');
                } catch(e) { showToast(e.message, 'error'); }
              }}>🔗 Share</button>
              <button id="btn-export-pdf" className="btn btn-outline" onClick={exportPDF} disabled={exporting || !sheet?.entries?.length}>
                {exporting ? <span className="spinner"/> : '↓ PDF'}
              </button>
              <button id="btn-export-excel" className="btn btn-outline" onClick={exportExcel} disabled={exporting || !sheet?.entries?.length}>
                {exporting ? <span className="spinner"/> : '↓ Excel'}
              </button>
              <button id="btn-add-item" className="btn btn-primary" onClick={() => setShowAdd(true)}>+ Add Item</button>
            </div>
          </div>

          {/* Sheets Tab Bar */}
          <div style={{ display:'flex', gap:8, marginBottom: 16, overflowX: 'auto', paddingBottom: 4 }}>
            {!loadingSheets && sheets.map(s => (
              <button 
                key={s.id}
                className={activeSheetID === s.id ? 'btn btn-primary btn-sm' : 'btn btn-outline btn-sm'} 
                style={{ borderRadius: '16px' }}
                onClick={() => setActiveSheetID(s.id)}
              >
                {s.name}
              </button>
            ))}
            <button className="btn btn-outline btn-sm" style={{ borderRadius: '16px', borderStyle: 'dashed' }} onClick={async () => {
              const name = prompt('New sheet name:');
              if (name) {
                try {
                  const s = await api('POST', '/api/projects/' + project.ID + '/sheets', { name });
                  setSheets(prev => [...prev, s]);
                  setActiveSheetID(s.id);
                  showToast('Sheet created!');
                } catch(e) { showToast(e.message, 'error'); }
              }
            }}>+ New Sheet</button>
          </div>

          <div className="card" style={{ padding:0, overflow:'hidden' }}>
            {loading ? (
              <div style={{ textAlign:'center', padding:60 }}><span className="spinner"/></div>
            ) : !sheet?.entries?.length ? (
              <div className="empty-state">
                <div className="icon">📋</div>
                <h3>BOQ sheet is empty</h3>
                <p>Add your first work item to get started</p>
                <button className="btn btn-primary" style={{ marginTop:16 }} onClick={() => setShowAdd(true)}>+ Add Item</button>
              </div>
            ) : (
              <div style={{ overflowX:'auto' }}>
                <table>
                  <thead>
                    <tr>
                      <th style={{ width:40 }}>Sr.</th>
                      <th>Description of Work</th>
                      <th style={{ width:110 }}>Category</th>
                      <th className="text-right" style={{ width:65 }}>L (m)</th>
                      <th className="text-right" style={{ width:65 }}>B (m)</th>
                      <th className="text-right" style={{ width:65 }}>H (m)</th>
                      <th className="text-right" style={{ width:65 }}>Unit</th>
                      <th className="text-right" style={{ width:90 }}>Quantity</th>
                      <th className="text-right" style={{ width:100 }}>Rate (₹)</th>
                      <th className="text-right" style={{ width:120 }}>Amount (₹)</th>
                      <th style={{ width:40 }}></th>
                    </tr>
                  </thead>
                  <tbody>
                    {(() => {
                      const rows = [];
                      let lastCat = null;
                      let sr = 1;

                      sheet.entries.forEach(e => {
                        if (e.Category !== lastCat) {
                          rows.push(
                            <tr key={'cat_' + e.Category} style={{ background: 'var(--surface2)' }}>
                              <td colSpan={11} style={{ fontWeight: 700, color: 'var(--text)', paddingTop: 14, paddingBottom: 14 }}>
                                Category: {e.Category}
                              </td>
                            </tr>
                          );
                          lastCat = e.Category;
                        }

                        const editMode = editingRow === e.ID;
                        if (editMode) {
                          rows.push(
                            <tr key={e.ID} style={{ background:'rgba(59, 130, 246, 0.15)' }}>
                              <td className="text-center" style={{ color:'var(--text-muted)' }}>{sr++}</td>
                              <td style={{ maxWidth:260 }}>
                                <div style={{ fontWeight:500 }}>{e.Description}</div>
                              </td>
                              <td><span className="badge badge-blue">{e.Category}</span></td>
                              <td className="text-right"><input type="number" step="0.001" className="input" style={{ width:66, padding:'4px', textAlign:'right' }} value={editForm.length} onChange={setEdit('length')} /></td>
                              <td className="text-right"><input type="number" step="0.001" className="input" style={{ width:66, padding:'4px', textAlign:'right' }} value={editForm.breadth} onChange={setEdit('breadth')} /></td>
                              <td className="text-right"><input type="number" step="0.001" className="input" style={{ width:66, padding:'4px', textAlign:'right' }} value={editForm.height} onChange={setEdit('height')} /></td>
                              <td className="text-center"><span className="tag">{e.Unit}</span></td>
                              <td className="text-right"><input type="number" step="0.001" className="input" style={{ width:76, padding:'4px', textAlign:'right' }} value={editForm.manual_qty} onChange={setEdit('manual_qty')} /></td>
                              <td className="text-right"><input type="number" step="0.01" className="input" style={{ width:86, padding:'4px', textAlign:'right' }} value={editForm.manual_rate} onChange={setEdit('manual_rate')} title="Base Rate (before index)" /></td>
                              <td className="text-right amount-cell" style={{ color:'var(--text-muted)' }}>—</td>
                              <td className="text-center" style={{ display:'flex', gap:4, justifyContent:'center' }}>
                                <button className="btn btn-primary btn-sm" style={{ padding:'3px' }} onClick={() => saveEdit(e.ID)} disabled={savingEdit}>✓</button>
                                <button className="btn btn-outline btn-sm" style={{ padding:'3px' }} onClick={() => setEditingRow(null)}>✕</button>
                              </td>
                            </tr>
                          );
                        } else {
                          rows.push(
                            <tr key={e.ID} onDoubleClick={() => startEdit(e)}>
                              <td className="text-center" style={{ color:'var(--text-muted)' }}>{sr++}</td>
                              <td style={{ maxWidth:260 }}>
                                <div style={{ fontWeight:500 }}>{e.Description}</div>
                              </td>
                              <td><span className="badge badge-blue">{e.Category}</span></td>
                              <td className="text-right" style={{ color:'var(--text-muted)', fontVariantNumeric:'tabular-nums', cursor:'pointer' }} onClick={() => startEdit(e)}>{e.Length>0?e.Length.toFixed(2):'—'}</td>
                              <td className="text-right" style={{ color:'var(--text-muted)', fontVariantNumeric:'tabular-nums', cursor:'pointer' }} onClick={() => startEdit(e)}>{e.Breadth>0?e.Breadth.toFixed(2):'—'}</td>
                              <td className="text-right" style={{ color:'var(--text-muted)', fontVariantNumeric:'tabular-nums', cursor:'pointer' }} onClick={() => startEdit(e)}>{e.Height>0?e.Height.toFixed(2):'—'}</td>
                              <td className="text-center"><span className="tag">{e.Unit}</span></td>
                              <td className="text-right amount-cell" style={{ cursor:'pointer' }} onClick={() => startEdit(e)}>{e.Quantity.toFixed(3)}</td>
                              <td className="text-right" style={{ cursor:'pointer' }} onClick={() => startEdit(e)}>{fmt(e.Rate)}</td>
                              <td className="text-right amount-cell" style={{ color:'var(--accent)', fontVariantNumeric:'tabular-nums' }}>{fmt(e.Amount)}</td>
                              <td className="text-center" style={{ display:'flex', gap:4, justifyContent:'center' }}>
                                <button className="btn btn-outline btn-sm" style={{ padding:'2px 6px' }} title="Edit inline" onClick={() => startEdit(e)}>✎</button>
                                <button className="btn btn-danger btn-sm" style={{ padding:'2px 6px' }} title="Remove" onClick={() => deleteEntry(e.ID)}>✕</button>
                              </td>
                            </tr>
                          );
                        }
                      });
                      return rows;
                    })()}
                    <tr className="grand-total-row">
                      <td colSpan={9} style={{ textAlign:'right', paddingRight:16 }}>Grand Total</td>
                      <td className="text-right amount-cell">₹{fmt(sheet.grand_total)}</td>
                      <td></td>
                    </tr>
                  </tbody>
                </table>
              </div>
            )}
          </div>

          {showAdd && (
            <BOQEntryModal
              projectID={project.ID}
              sheetID={activeSheetID}
              showToast={showToast}
              onClose={() => setShowAdd(false)}
              onSaved={() => { setShowAdd(false); load(); }}
            />
          )}
        </div>
      );
    }

    }

    // ─── Public Share View ────────────────────────────────────────────────────────
    function ShareView({ token, showToast }) {
      const [sheet, setSheet] = useState(null);
      const [loading, setLoading] = useState(true);

      useEffect(() => {
        api('GET', '/api/share/' + token)
          .then(setSheet)
          .catch(e => showToast(e.message, 'error'))
          .finally(() => setLoading(false));
      }, [token]);

      const fmt = (n) => (n||0).toLocaleString('en-IN', { minimumFractionDigits:2, maximumFractionDigits:2 });

      if (loading) return <div style={{ padding:60, textAlign:'center' }}><span className="spinner"/></div>;
      if (!sheet) return <div style={{ padding:60, textAlign:'center' }}>View Not Found.</div>;

      let sr = 1;
      return (
        <div style={{ maxWidth: 1000, margin: '40px auto', padding: 20 }}>
          <div className="card" style={{ padding: 24, marginBottom: 20, textAlign: 'center' }}>
            <h2 style={{ margin:0, marginBottom:8 }}>{sheet.project.Name}</h2>
            <p style={{ margin:0, color:'var(--text-muted)' }}>
              {sheet.project.ClientName && <span>👤 Client: {sheet.project.ClientName} &bull; </span>}
              📍 Location: {sheet.project.Location}
            </p>
          </div>
          <div className="card" style={{ overflowX: 'auto', padding: 0 }}>
            <table className="table" style={{ width: '100%', minWidth: 800 }}>
              <thead>
                <tr>
                  <th style={{ width: 50, textAlign:'center' }}>Sr</th>
                  <th>Description</th>
                  <th>Category</th>
                  <th className="text-right">L</th>
                  <th className="text-right">B</th>
                  <th className="text-right">H</th>
                  <th className="text-center">Unit</th>
                  <th className="text-right">Qty</th>
                  <th className="text-right">Rate</th>
                  <th className="text-right" style={{ paddingRight:16 }}>Amount</th>
                </tr>
              </thead>
              <tbody>
                {sheet.entries?.map(e => (
                  <tr key={e.ID}>
                    <td className="text-center" style={{ color: 'var(--text-muted)' }}>{sr++}</td>
                    <td><div style={{ fontWeight: 500 }}>{e.Description}</div></td>
                    <td><span className="badge badge-blue">{e.Category}</span></td>
                    <td className="text-right" style={{ color: 'var(--text-muted)' }}>{e.Length > 0 ? e.Length.toFixed(2) : '-'}</td>
                    <td className="text-right" style={{ color: 'var(--text-muted)' }}>{e.Breadth > 0 ? e.Breadth.toFixed(2) : '-'}</td>
                    <td className="text-right" style={{ color: 'var(--text-muted)' }}>{e.Height > 0 ? e.Height.toFixed(2) : '-'}</td>
                    <td className="text-center"><span className="tag">{e.Unit}</span></td>
                    <td className="text-right amount-cell">{e.Quantity.toFixed(3)}</td>
                    <td className="text-right">{fmt(e.Rate)}</td>
                    <td className="text-right amount-cell" style={{ color: 'var(--accent)' }}>{fmt(e.Amount)}</td>
                  </tr>
                ))}
                {!sheet.entries?.length && (
                  <tr><td colSpan={10} style={{ textAlign: 'center', padding: 40, color: 'var(--text-muted)' }}>This project is empty.</td></tr>
                )}
                <tr className="grand-total-row">
                  <td colSpan={9} style={{ textAlign: 'right', paddingRight: 16 }}>GRAND TOTAL</td>
                  <td className="text-right amount-cell" style={{ fontSize: 16 }}>₹{fmt(sheet.grand_total)}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      );
    }

    // ─── App Root ─────────────────────────────────────────────────────────────────
    function App() {
      const p = window.location.pathname;
      const [toastEl, showToast] = useToast();

      const [user, setUser] = useState(() => {
        // Handle Google OAuth callback: pick up ?token= from URL
        const params = new URLSearchParams(window.location.search);
        const urlToken = params.get('token');
        if (urlToken) {
          // Decode user from JWT payload (middle base64 part)
          try {
            const payload = JSON.parse(atob(urlToken.split('.')[1]));
            const u = { id: payload.sub, name: payload.name, email: '' };
            localStorage.setItem('xtm_token', urlToken);
            localStorage.setItem('xtm_user', JSON.stringify(u));
            // Clean URL
            window.history.replaceState({}, '', '/');
            return u;
          } catch(e) { /* ignore malformed token */ }
        }
        return getUser();
      });
      const [page, setPage] = useState('dashboard');
      const [currentProject, setCurrentProject] = useState(null);
      const [projects, setProjects] = useState([]);
      const [projectsLoading, setProjectsLoading] = useState(false);

      const loadProjects = useCallback(async () => {
        if (!getToken()) return;
        setProjectsLoading(true);
        try { setProjects(await api('GET', '/api/projects') || []); }
        catch(e) { console.error('Failed to load projects', e); }
        finally { setProjectsLoading(false); }
      }, []);

      useEffect(() => {
        if (user && !p.startsWith('/share/')) loadProjects();
      }, [user, loadProjects, p]);

      const login = (u) => { setUser(u); setPage('dashboard'); };
      const logout = () => {
        localStorage.removeItem('xtm_token');
        localStorage.removeItem('xtm_user');
        setUser(null); setPage('dashboard'); setCurrentProject(null);
      };
      const openProject = (p) => { setCurrentProject(p); setPage('project'); };
      const backToDash = () => { setCurrentProject(null); setPage('dashboard'); };

      if (p.startsWith('/share/')) {
        const token = p.split('/share/')[1];
        return <><ShareView token={token} showToast={showToast} />{toastEl}</>;
      }

      if (!user) return <><AuthPage onLogin={login}/>{toastEl}</>;

      return (
        <>
          <Sidebar user={user} page={page} onNav={setPage} onLogout={logout} projects={projects} currentProject={currentProject} onOpenProject={openProject} />
          <div className="main-content">
            {page === 'dashboard' && <Dashboard projects={projects} loading={projectsLoading} onRefresh={loadProjects} onOpenProject={openProject} showToast={showToast}/>}
            {page === 'project' && currentProject && <ProjectSheet project={currentProject} onBack={backToDash} showToast={showToast}/>}
          </div>
          {toastEl}
        </>
      );
    }

    ReactDOM.createRoot(document.getElementById('root')).render(<App/>);
