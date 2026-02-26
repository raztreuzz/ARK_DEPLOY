import React, { useState, useEffect, useMemo } from 'react';
import {
  Search, RefreshCw, Plus, MoreVertical, Terminal,
  Trash2, ExternalLink, Activity, Package, AlertCircle,
  CheckCircle2, Globe, Cpu, X, Loader2
} from 'lucide-react';

// --- CONSTANTES Y CONFIGURACION ---
const ENVS = ['PROD', 'DEV', 'TEST'];
const INSTANCE_STATUS = {
  RUNNING: 'running',
  FAILED: 'failed',
  PROVISIONING: 'provisioning',
  STOPPED: 'stopped'
};

const dbg = (...args) => console.log('[Admin]', ...args);

// --- HOOK PRINCIPAL DE ESTADO ---
function useAdminData() {
  const [products, setProducts] = useState([]);
  const [instances, setInstances] = useState([]);
  const [devices, setDevices] = useState([]);
  const [sshUsers, setSSHUsers] = useState({});
  const [jobsCatalog, setJobsCatalog] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      dbg('fetchData start');
      const [pRes, iRes, dRes] = await Promise.all([
        fetch('/api/products').then((r) => (r.ok ? r.json() : Promise.reject('Error en productos'))),
        fetch('/api/deployments').then((r) => (r.ok ? r.json() : Promise.reject('Error en instancias'))),
        fetch('/api/tailscale/devices').then((r) => (r.ok ? r.json() : { devices: [] })).catch(() => ({ devices: [] }))
      ]);
      const derivedJobs = new Set();
      (pRes.products || []).forEach((p) => {
        Object.values(p.deploy_jobs || {}).forEach((j) => {
          const v = String(j || '').trim();
          if (v) derivedJobs.add(v);
        });
        const del = String(p.delete_job || '').trim();
        if (del) derivedJobs.add(del);
      });
      const sshRes = await fetch('/api/ssh-users').then((r) => (r.ok ? r.json() : { map: {} })).catch(() => ({ map: {} }));
      dbg('products loaded', pRes.products?.length || 0);
      dbg('instances loaded', iRes.instances?.length || 0);
      dbg('devices loaded', dRes.devices?.length || 0);
      dbg('jobs loaded (derived)', derivedJobs.size);
      dbg('ssh users loaded', Object.keys(sshRes.map || {}).length);
      setProducts(pRes.products || []);
      setInstances(iRes.instances || []);
      setDevices(dRes.devices || []);
      setSSHUsers(sshRes.map || {});
      setJobsCatalog(Array.from(derivedJobs));
    } catch (err) {
      console.error('[Admin] fetchData error', err);
      setError(err.toString());
    } finally {
      dbg('fetchData done');
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, []);

  return { products, instances, devices, sshUsers, setSSHUsers, jobsCatalog, loading, error, fetchData };
}

// --- COMPONENTE PRINCIPAL (Layout) ---
export default function AdminDashboard() {
  const { products, instances, devices, sshUsers, setSSHUsers, jobsCatalog, loading, error, fetchData } = useAdminData();
  const [filter, setFilter] = useState('');
  const [modals, setModals] = useState({ product: null, productDelete: null, logs: null, delete: null });

  const jobOptions = useMemo(() => {
    const set = new Set();
    (jobsCatalog || []).forEach((job) => {
      const v = typeof job === 'string' ? job : (job?.name || job?.id || '');
      const normalized = String(v || '').trim();
      if (normalized) set.add(normalized);
    });
    products.forEach((p) => {
      Object.values(p.deploy_jobs || {}).forEach((job) => {
        const v = String(job || '').trim();
        if (v) set.add(v);
      });
      const del = String(p.delete_job || '').trim();
      if (del) set.add(del);
    });
    return Array.from(set).sort((a, b) => a.localeCompare(b));
  }, [products, jobsCatalog]);

  const filteredProducts = products.filter((p) =>
    p.name.toLowerCase().includes(filter.toLowerCase()) ||
    p.id.toLowerCase().includes(filter.toLowerCase())
  );

  const handleSaveProduct = async (formData) => {
    const payload = {
      id: formData.id?.trim(),
      name: formData.name?.trim(),
      description: formData.description?.trim() || '',
      deploy_jobs: formData.deploy_jobs || {},
      delete_job: formData.delete_job?.trim() || '',
      web_service: formData.web_service_enabled ? (formData.web_service?.trim() || 'web') : '',
      web_port: Number(formData.web_port) || 80
    };

    const isEdit = Boolean(modals.product?.id);
    const url = isEdit ? `/api/products/${modals.product.id}` : '/api/products';
    const method = isEdit ? 'PUT' : 'POST';
    dbg('save product', method, url, payload);

    const res = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    dbg('save product status', res.status);

    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      console.error('[Admin] save product error response', err);
      throw new Error(err.detail || res.statusText);
    }

    dbg('save product success');
    setModals((m) => ({ ...m, product: null }));
    await fetchData();
  };

  const handleDeleteInstance = async (instance) => {
    dbg('delete instance', instance?.id);
    const res = await fetch(`/api/deployments/${instance.id}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' }
    });
    dbg('delete instance status', res.status);

    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      console.error('[Admin] delete instance error response', err);
      throw new Error(err.detail || res.statusText);
    }

    dbg('delete instance success');
    setModals((m) => ({ ...m, delete: null }));
    await fetchData();
  };

  const handleDeleteProduct = async (product) => {
    dbg('delete product', product?.id);
    const res = await fetch(`/api/products/${product.id}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' }
    });
    dbg('delete product status', res.status);

    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      console.error('[Admin] delete product error response', err);
      throw new Error(err.detail || res.statusText);
    }

    dbg('delete product success');
    setModals((m) => ({ ...m, productDelete: null }));
    await fetchData();
  };

  const handleSaveNodeSSHUser = async (host, sshUser) => {
    const normalizedHost = String(host || '').trim();
    const normalizedUser = String(sshUser || '').trim();
    if (!normalizedHost || !normalizedUser) {
      throw new Error('Host y ssh_user son requeridos');
    }
    const res = await fetch(`/api/ssh-users/${encodeURIComponent(normalizedHost)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ssh_user: normalizedUser })
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.detail || res.statusText);
    }
    setSSHUsers((prev) => ({ ...prev, [normalizedHost]: normalizedUser }));
  };

  if (loading && products.length === 0) return <LoadingState />;

  return (
    <div className="min-h-screen bg-[#020617] text-slate-200 font-sans">
      <TopBar filter={filter} setFilter={setFilter} onRefresh={fetchData} />

      <main className="max-w-7xl mx-auto p-4 md:p-8 space-y-8">
        {error && <ApiErrorBanner message={error} />}

        <section className="space-y-4">
          <SectionHeader
            title="Catalogo de Productos"
            icon={<Package size={18} />}
            action={(
              <button
                onClick={() => setModals((m) => ({ ...m, product: {} }))}
                className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-xl text-xs font-bold transition-all"
              >
                <Plus size={16} /> Nuevo Producto
              </button>
            )}
          />
          <ProductsTable
            products={filteredProducts}
            onEdit={(p) => setModals((m) => ({ ...m, product: p }))}
            onDelete={(p) => setModals((m) => ({ ...m, productDelete: p }))}
          />
        </section>

        <section className="space-y-4">
          <SectionHeader title="Instancias en Ejecucion" icon={<Activity size={18} />} />
          <InstanceList
            instances={instances}
            onViewLogs={(id) => setModals((m) => ({ ...m, logs: id }))}
            onDelete={(inst) => setModals((m) => ({ ...m, delete: inst }))}
          />
        </section>

        <section className="space-y-4">
          <SectionHeader title="Nodos Conectados" icon={<Globe size={18} />} />
          <TailscaleNodesPanel devices={devices} sshUsers={sshUsers} onSaveSSHUser={handleSaveNodeSSHUser} />
        </section>

        <section className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <div className="lg:col-span-12">
            <SystemLogConsole />
          </div>
        </section>
      </main>

      {modals.product !== null && (
        <ProductFormModal
          product={modals.product}
          jobOptions={jobOptions}
          onClose={() => setModals((m) => ({ ...m, product: null }))}
          onSave={handleSaveProduct}
        />
      )}
      {modals.logs && (
        <InstanceLogsModal
          instanceId={modals.logs}
          onClose={() => setModals((m) => ({ ...m, logs: null }))}
        />
      )}
      {modals.delete && (
        <InstanceDeleteConfirmModal
          instance={modals.delete}
          onClose={() => setModals((m) => ({ ...m, delete: null }))}
          onConfirm={() => handleDeleteInstance(modals.delete)}
        />
      )}
      {modals.productDelete && (
        <ProductDeleteConfirmModal
          product={modals.productDelete}
          onClose={() => setModals((m) => ({ ...m, productDelete: null }))}
          onConfirm={() => handleDeleteProduct(modals.productDelete)}
        />
      )}

      <CustomScrollbarStyles />
    </div>
  );
}

// --- COMPONENTES FASE 1 ---

const TopBar = ({ filter, setFilter, onRefresh }) => (
  <nav className="sticky top-0 z-40 bg-[#020617]/80 backdrop-blur-md border-b border-slate-800 px-6 py-4">
    <div className="max-w-7xl mx-auto flex items-center justify-between gap-4">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center font-black text-white italic">A</div>
        <h1 className="text-sm font-black uppercase tracking-tighter hidden md:block">Ark <span className="text-blue-500">Admin</span></h1>
      </div>

      <div className="flex-1 max-w-md relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={16} />
        <input
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder="Buscar productos o IDs..."
          className="w-full bg-slate-900 border border-slate-800 rounded-xl py-2 pl-10 pr-4 text-xs focus:ring-1 ring-blue-500 outline-none transition-all"
        />
      </div>

      <button onClick={onRefresh} className="p-2 hover:bg-slate-800 rounded-xl text-slate-400 transition-colors">
        <RefreshCw size={18} />
      </button>
    </div>
  </nav>
);

const ProductsTable = ({ products, onEdit, onDelete }) => (
  <div className="bg-slate-900/40 border border-slate-800 rounded-2xl overflow-hidden">
    <table className="w-full text-left border-collapse">
      <thead>
        <tr className="bg-slate-800/50 text-[10px] uppercase font-black tracking-widest text-slate-500">
          <th className="px-6 py-4">Producto</th>
          <th className="px-6 py-4">Configuracion Web</th>
          <th className="px-6 py-4">Jobs de Pipeline</th>
          <th className="px-6 py-4 text-right">Acciones</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-slate-800/50">
        {products.map((p) => (
          <tr key={p.id} className="hover:bg-slate-800/20 transition-colors group">
            <td className="px-6 py-4">
              <div className="font-bold text-slate-100">{p.name}</div>
              <div className="text-[10px] text-slate-500 font-mono">{p.id}</div>
            </td>
            <td className="px-6 py-4">
              {p.web_service ? (
                <div className="flex items-center gap-2">
                  <Globe size={12} className="text-blue-400" />
                  <span className="text-xs text-slate-300">{p.web_service}:{p.web_port || 80}</span>
                </div>
              ) : <span className="text-[10px] text-slate-600 italic">No Web UI</span>}
            </td>
            <td className="px-6 py-4">
              <div className="flex gap-1">
                {Object.keys(p.deploy_jobs || {}).map((env) => (
                  <span key={env} className="text-[8px] px-1.5 py-0.5 bg-slate-800 rounded text-slate-400 border border-slate-700">{env}</span>
                ))}
              </div>
            </td>
            <td className="px-6 py-4 text-right">
              <button onClick={() => onEdit(p)} className="p-2 hover:bg-slate-700 rounded-lg text-slate-400 mr-2" title="Editar producto">
                <MoreVertical size={16} />
              </button>
              <button onClick={() => onDelete(p)} className="p-2 hover:bg-red-500/20 rounded-lg text-slate-400 hover:text-red-400" title="Eliminar producto">
                <Trash2 size={16} />
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
    {products.length === 0 && <EmptyState message="No se encontraron productos." />}
  </div>
);

const InstanceList = ({ instances, onViewLogs, onDelete }) => (
  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
    {instances.map((inst) => (
      <div key={inst.id} className="bg-slate-900/60 border border-slate-800 rounded-2xl p-5 hover:border-slate-700 transition-all">
        <div className="flex justify-between items-start mb-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-600/10 rounded-lg border border-blue-500/20">
              <Cpu size={18} className="text-blue-400" />
            </div>
            <div>
              <h4 className="text-xs font-black uppercase text-slate-100">{inst.product_id}</h4>
              <p className="text-[10px] text-slate-500 font-mono">{inst.device_id}</p>
            </div>
          </div>
          <StatusBadge status={inst.status} />
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between text-[10px]">
            <span className="text-slate-500 uppercase font-bold tracking-widest">Entorno</span>
            <span className={`px-2 py-0.5 rounded font-bold ${String(inst.environment).toLowerCase() === 'prod' ? 'bg-red-500/10 text-red-500' : 'bg-yellow-500/10 text-yellow-500'}`}>{inst.environment}</span>
          </div>
          {inst.url && (
            <a href={inst.url} target="_blank" rel="noreferrer" className="flex items-center justify-between text-[10px] p-2 bg-black/40 rounded-lg border border-slate-800 hover:border-blue-500/50 group transition-all">
              <span className="text-slate-400 truncate mr-2">{inst.url}</span>
              <ExternalLink size={12} className="text-slate-600 group-hover:text-blue-400" />
            </a>
          )}
        </div>

        <div className="mt-5 pt-4 border-t border-slate-800/50 flex gap-2">
          <button onClick={() => onViewLogs(inst.id)} className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 rounded-xl text-[10px] font-black uppercase tracking-widest transition-colors flex items-center justify-center gap-2">
            <Terminal size={12} /> Logs
          </button>
          <button onClick={() => onDelete(inst)} className="p-2 bg-red-500/10 hover:bg-red-500 text-red-500 hover:text-white rounded-xl transition-all border border-red-500/20">
            <Trash2 size={16} />
          </button>
        </div>
      </div>
    ))}
    {instances.length === 0 && <div className="col-span-full py-12 border-2 border-dashed border-slate-800 rounded-3xl text-center text-slate-600 text-sm">No hay instancias activas.</div>}
  </div>
);

const TailscaleNodesPanel = ({ devices, sshUsers, onSaveSSHUser }) => {
  const [draftUsers, setDraftUsers] = useState({});
  const [savingHost, setSavingHost] = useState('');
  const [saveError, setSaveError] = useState('');

  const getDraft = (ip) => {
    if (Object.prototype.hasOwnProperty.call(draftUsers, ip)) return draftUsers[ip];
    return sshUsers[ip] || '';
  };

  const handleSave = async (ip) => {
    setSavingHost(ip);
    setSaveError('');
    try {
      await onSaveSSHUser(ip, getDraft(ip));
    } catch (e) {
      setSaveError(e.message || 'Error guardando usuario SSH');
    } finally {
      setSavingHost('');
    }
  };

  return (
    <div className="bg-slate-900/40 border border-slate-800 rounded-2xl overflow-hidden">
      <div className="bg-slate-800/40 px-4 py-3 border-b border-slate-800 flex items-center justify-between">
        <span className="text-xs font-bold text-slate-300">Dispositivos de la malla</span>
        <span className="text-[10px] text-slate-500 font-mono">{devices.length} nodo(s)</span>
      </div>
      {saveError && <div className="mx-4 mt-3 text-xs text-red-400">{saveError}</div>}
      <div className="divide-y divide-slate-800/60">
        {devices.map((d, idx) => {
          const host = d.hostname || d.name || `device-${idx + 1}`;
          const normalize = (v) => String(v || '').trim().split('/')[0];
          const ip = (Array.isArray(d.addresses) && d.addresses.length > 0 ? normalize(d.addresses[0]) : normalize(d.ip)) || 'N/A';
          const status = String(d.status || '').toLowerCase();
          const state = String(d.state || '').toLowerCase();
          const hasTSIP = /^100\.\d{1,3}\.\d{1,3}\.\d{1,3}$/.test(ip);
          const online = d.online === true || d.active === true || status === 'active' || status === 'online' || state === 'active' || state === 'online';
          const lastSeen = d.lastSeen ? new Date(d.lastSeen).getTime() : 0;
          const recentlySeen = lastSeen > 0 && (Date.now() - lastSeen) < 10 * 60 * 1000;
          const reachable = !online && hasTSIP && recentlySeen;
          const badgeClass = online
            ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
            : reachable
              ? 'bg-blue-500/10 text-blue-400 border-blue-500/30'
              : 'bg-slate-800 text-slate-500 border-slate-700';
          const badgeText = online ? 'Online' : (reachable ? 'Reachable' : 'Offline');
          return (
            <div key={`${host}-${ip}-${idx}`} className="px-4 py-3 space-y-3">
              <div className="flex items-center justify-between">
                <div>
                  <div className="text-sm text-slate-200 font-semibold">{host}</div>
                  <div className="text-[10px] text-slate-500 font-mono">
                    {ip}{d.lastSeen ? ` · seen ${new Date(d.lastSeen).toLocaleTimeString()}` : ''}
                  </div>
                </div>
                <span className={`px-2 py-1 rounded-full text-[9px] font-black uppercase tracking-widest border ${badgeClass}`}>
                  {badgeText}
                </span>
              </div>
              {hasTSIP && (
                <div className="flex items-center gap-2">
                  <input
                    value={getDraft(ip)}
                    onChange={(e) => setDraftUsers((prev) => ({ ...prev, [ip]: e.target.value }))}
                    placeholder="ssh user (ej: SARA, ubuntu, root)"
                    className="flex-1 bg-slate-900 border border-slate-800 rounded-lg px-3 py-2 text-xs outline-none"
                  />
                  <button
                    type="button"
                    onClick={() => handleSave(ip)}
                    disabled={savingHost === ip}
                    className="px-3 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-60 text-white text-[10px] font-black uppercase tracking-widest"
                  >
                    {savingHost === ip ? 'Guardando' : 'Guardar'}
                  </button>
                </div>
              )}
            </div>
          );
        })}
        {devices.length === 0 && (
          <div className="px-4 py-8 text-center text-slate-600 text-xs">
            No se detectaron nodos conectados.
          </div>
        )}
      </div>
    </div>
  );
};

// --- MODALES (FASE 1) ---

const ProductFormModal = ({ product, jobOptions, onClose, onSave }) => {
  const [form, setForm] = useState(product?.id ? {
    ...product,
    deploy_jobs: {
      PROD: product.deploy_jobs?.PROD || product.deploy_jobs?.prod || '',
      DEV: product.deploy_jobs?.DEV || product.deploy_jobs?.dev || '',
      TEST: product.deploy_jobs?.TEST || product.deploy_jobs?.test || ''
    },
    web_service_enabled: Boolean(product.web_service),
    web_service: product.web_service || 'web',
    web_port: product.web_port || 80
  } : {
    id: '', name: '', description: '',
    deploy_jobs: { PROD: '', DEV: '', TEST: '' },
    delete_job: '',
    web_service_enabled: false,
    web_service: 'web',
    web_port: 80
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const hasJobs = (jobOptions || []).length > 0;

  const handleSubmit = async () => {
    setSaving(true);
    setError('');
    try {
      await onSave(form);
    } catch (e) {
      setError(e.message || 'Error guardando producto');
      setSaving(false);
      return;
    }
    setSaving(false);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-[#020617]/90 backdrop-blur-sm">
      <div className="bg-slate-950 border border-slate-800 w-full max-w-2xl rounded-[2.5rem] shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <div className="px-8 py-6 border-b border-slate-800 flex items-center justify-between">
          <h2 className="text-xs font-black uppercase tracking-widest flex items-center gap-2">
            <Package size={16} className="text-blue-500" /> {product?.id ? 'Editar Producto' : 'Nuevo Producto'}
          </h2>
          <button onClick={onClose} className="text-slate-500 hover:text-white"><X size={20} /></button>
        </div>

        <div className="p-8 overflow-y-auto space-y-6 custom-scrollbar">
          <div className="grid grid-cols-2 gap-4">
            <FormField label="ID unico (Slug)" value={form.id} onChange={(v) => setForm({ ...form, id: v })} placeholder="ej: my-app-api" disabled={!!product?.id} />
            <FormField label="Nombre Comercial" value={form.name} onChange={(v) => setForm({ ...form, name: v })} placeholder="ej: API de Usuarios" />
          </div>

          <div className="space-y-1">
            <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Descripcion</label>
            <textarea
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="w-full bg-slate-900 border border-slate-800 rounded-xl p-3 text-xs outline-none focus:ring-1 ring-blue-500 h-20"
            />
          </div>

          <div className="grid grid-cols-2 gap-6">
            <div className="space-y-4">
              <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Configuracion Web</label>
              <div className="flex items-center gap-4">
                <button
                  onClick={() => setForm({ ...form, web_service_enabled: !form.web_service_enabled })}
                  className={`relative w-10 h-5 rounded-full transition-colors ${form.web_service_enabled ? 'bg-blue-600' : 'bg-slate-800'}`}
                >
                  <div className={`absolute top-1 w-3 h-3 bg-white rounded-full transition-all ${form.web_service_enabled ? 'left-6' : 'left-1'}`} />
                </button>
                <span className="text-xs text-slate-300">¿Tiene Interfaz Web?</span>
              </div>
              {form.web_service_enabled && (
                <>
                  <FormField label="Servicio Web" value={form.web_service} onChange={(v) => setForm({ ...form, web_service: v })} placeholder="web" />
                  <FormField label="Puerto" type="number" value={form.web_port} onChange={(v) => setForm({ ...form, web_port: v })} />
                </>
              )}
            </div>

            <div className="space-y-3">
              <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Jobs de Despliegue</label>
              {ENVS.map((env) => (
                <div key={env} className="flex items-center gap-2">
                  <span className="text-[10px] w-10 font-bold text-slate-600">{env}</span>
                  <select
                    value={form.deploy_jobs[env] || ''}
                    onChange={(e) => setForm({ ...form, deploy_jobs: { ...form.deploy_jobs, [env]: e.target.value } })}
                    className="flex-1 bg-slate-900 border border-slate-800 rounded-lg px-3 py-1.5 text-[10px] outline-none text-slate-200"
                    disabled={!hasJobs}
                  >
                    <option value="">Seleccionar job...</option>
                    {(jobOptions || []).map((job) => (
                      <option key={`${env}-${job}`} value={job}>{job}</option>
                    ))}
                  </select>
                </div>
              ))}
              <div className="pt-2">
                <div className="space-y-1">
                  <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Job de Eliminacion</label>
                  <select
                    value={form.delete_job}
                    onChange={(e) => setForm({ ...form, delete_job: e.target.value })}
                    className="w-full bg-slate-900 border border-slate-800 rounded-xl px-4 py-2.5 text-xs outline-none focus:ring-1 ring-blue-500 transition-all text-slate-200"
                    disabled={!hasJobs}
                  >
                    <option value="">Seleccionar job...</option>
                    {(jobOptions || []).map((job) => (
                      <option key={`delete-${job}`} value={job}>{job}</option>
                    ))}
                  </select>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-2">
            <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Biblioteca de Jobs</label>
            {hasJobs ? (
              <div className="max-h-40 overflow-y-auto custom-scrollbar border border-slate-800 rounded-xl p-3 space-y-2 bg-slate-900/40">
                {(jobOptions || []).map((job) => (
                  <div key={`quick-${job}`} className="flex items-center justify-between gap-2">
                    <span className="text-[10px] text-slate-300 font-mono truncate">{job}</span>
                    <div className="flex items-center gap-1">
                      <button type="button" onClick={() => setForm({ ...form, deploy_jobs: { ...form.deploy_jobs, PROD: job } })} className="px-2 py-0.5 rounded bg-slate-800 hover:bg-blue-600 text-[9px] font-bold">PROD</button>
                      <button type="button" onClick={() => setForm({ ...form, deploy_jobs: { ...form.deploy_jobs, DEV: job } })} className="px-2 py-0.5 rounded bg-slate-800 hover:bg-blue-600 text-[9px] font-bold">DEV</button>
                      <button type="button" onClick={() => setForm({ ...form, deploy_jobs: { ...form.deploy_jobs, TEST: job } })} className="px-2 py-0.5 rounded bg-slate-800 hover:bg-blue-600 text-[9px] font-bold">TEST</button>
                      <button type="button" onClick={() => setForm({ ...form, delete_job: job })} className="px-2 py-0.5 rounded bg-slate-800 hover:bg-red-600 text-[9px] font-bold">DEL</button>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-xs text-slate-500 border border-dashed border-slate-800 rounded-xl p-3">
                No hay jobs disponibles en el catalogo actual de productos.
              </div>
            )}
          </div>

          {error && <div className="text-xs text-red-400 bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-2">{error}</div>}
        </div>

        <div className="p-8 bg-slate-900/40 border-t border-slate-800 flex gap-4">
          <button onClick={onClose} className="flex-1 py-3 text-xs font-black uppercase text-slate-500 tracking-widest hover:text-slate-300 transition-colors">Cancelar</button>
          <button onClick={handleSubmit} disabled={saving} className="flex-[2] py-3 bg-blue-600 hover:bg-blue-500 disabled:opacity-60 text-white rounded-2xl text-xs font-black uppercase tracking-[0.2em] shadow-lg shadow-blue-600/20 transition-all">
            {saving ? 'Guardando...' : 'Guardar Producto'}
          </button>
        </div>
      </div>
    </div>
  );
};

const InstanceLogsModal = ({ instanceId, onClose }) => {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    dbg('logs modal fetch', instanceId);
    fetch(`/api/deployments/${instanceId}/logs`)
      .then((r) => r.json())
      .then((data) => {
        dbg('logs modal response keys', Object.keys(data.logs || {}));
        const entries = Object.entries(data.logs || {});
        const lines = [];
        entries.forEach(([job, content]) => {
          lines.push(`=== ${job} ===`);
          String(content || '').split('\n').forEach((line) => lines.push(line));
          lines.push('');
        });
        setLogs(lines);
        setLoading(false);
      })
      .catch((e) => {
        console.error('[Admin] logs modal fetch error', e);
        setLoading(false);
      });
  }, [instanceId]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-[#020617]/95 backdrop-blur-md">
      <div className="bg-black border border-slate-800 w-full max-w-4xl h-[80vh] rounded-3xl shadow-2xl flex flex-col overflow-hidden">
        <div className="p-5 border-b border-slate-800 flex items-center justify-between bg-slate-900/50">
          <div className="flex items-center gap-3">
            <Terminal size={18} className="text-blue-500" />
            <span className="text-xs font-black uppercase tracking-widest">Logs de Instancia: <span className="text-blue-400 font-mono font-normal tracking-normal">{instanceId}</span></span>
          </div>
          <button onClick={onClose} className="text-slate-500 hover:text-white"><X size={20} /></button>
        </div>
        <div className="flex-1 p-6 font-mono text-[11px] overflow-y-auto custom-scrollbar bg-black text-slate-300 selection:bg-blue-500/30">
          {loading ? <div className="flex items-center gap-2 text-slate-600"><Loader2 size={14} className="animate-spin" /> Cargando stream...</div> : (
            logs.map((l, idx) => <div key={idx} className="mb-1 leading-relaxed opacity-90 border-l border-slate-900 pl-4 hover:border-blue-500/30 transition-colors"><span className="text-slate-700 mr-4">{(idx + 1).toString().padStart(3, '0')}</span>{l}</div>)
          )}
          {!loading && logs.length === 0 && <div className="text-slate-700 italic">No hay logs disponibles para esta instancia.</div>}
        </div>
      </div>
    </div>
  );
};

const InstanceDeleteConfirmModal = ({ instance, onClose, onConfirm }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleConfirm = async () => {
    setLoading(true);
    setError('');
    try {
      await onConfirm();
    } catch (e) {
      setError(e.message || 'Error eliminando instancia');
      setLoading(false);
      return;
    }
    setLoading(false);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-red-950/20 backdrop-blur-sm">
      <div className="bg-slate-950 border border-red-500/20 w-full max-w-md rounded-[2rem] shadow-2xl p-8 space-y-6">
        <div className="w-16 h-16 bg-red-500/10 rounded-2xl flex items-center justify-center mx-auto text-red-500">
          <AlertCircle size={32} />
        </div>
        <div className="text-center space-y-2">
          <h3 className="text-lg font-black uppercase tracking-tighter">¿Eliminar Instancia?</h3>
          <p className="text-xs text-slate-500 leading-relaxed">Estas a punto de ejecutar el job de eliminacion para <span className="font-bold text-slate-300">{instance.product_id}</span> en <span className="font-bold text-slate-300">{instance.device_id}</span>. Esta accion no se puede deshacer.</p>
        </div>
        {error && <div className="text-xs text-red-400 bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-2">{error}</div>}
        <div className="flex gap-3">
          <button onClick={onClose} className="flex-1 py-3 text-xs font-bold text-slate-500 hover:text-slate-300">Cancelar</button>
          <button onClick={handleConfirm} disabled={loading} className="flex-[2] py-3 bg-red-600 hover:bg-red-500 disabled:opacity-60 text-white rounded-xl text-xs font-black uppercase tracking-widest shadow-lg shadow-red-600/20 transition-all">{loading ? 'Eliminando...' : 'Eliminar Ahora'}</button>
        </div>
      </div>
    </div>
  );
};

const ProductDeleteConfirmModal = ({ product, onClose, onConfirm }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleConfirm = async () => {
    setLoading(true);
    setError('');
    try {
      await onConfirm();
    } catch (e) {
      setError(e.message || 'Error eliminando producto');
      setLoading(false);
      return;
    }
    setLoading(false);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-red-950/20 backdrop-blur-sm">
      <div className="bg-slate-950 border border-red-500/20 w-full max-w-md rounded-[2rem] shadow-2xl p-8 space-y-6">
        <div className="w-16 h-16 bg-red-500/10 rounded-2xl flex items-center justify-center mx-auto text-red-500">
          <AlertCircle size={32} />
        </div>
        <div className="text-center space-y-2">
          <h3 className="text-lg font-black uppercase tracking-tighter">¿Eliminar Producto?</h3>
          <p className="text-xs text-slate-500 leading-relaxed">
            Vas a eliminar <span className="font-bold text-slate-300">{product.name}</span> ({product.id}).
            Esta acción no se puede deshacer.
          </p>
        </div>
        {error && <div className="text-xs text-red-400 bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-2">{error}</div>}
        <div className="flex gap-3">
          <button onClick={onClose} className="flex-1 py-3 text-xs font-bold text-slate-500 hover:text-slate-300">Cancelar</button>
          <button onClick={handleConfirm} disabled={loading} className="flex-[2] py-3 bg-red-600 hover:bg-red-500 disabled:opacity-60 text-white rounded-xl text-xs font-black uppercase tracking-widest shadow-lg shadow-red-600/20 transition-all">
            {loading ? 'Eliminando...' : 'Eliminar Producto'}
          </button>
        </div>
      </div>
    </div>
  );
};

// --- COMPONENTES AUXILIARES ---

const StatusBadge = ({ status }) => {
  const cfg = {
    [INSTANCE_STATUS.RUNNING]: { color: 'bg-green-500/10 text-green-500', icon: <CheckCircle2 size={10} />, label: 'Running' },
    [INSTANCE_STATUS.FAILED]: { color: 'bg-red-500/10 text-red-500', icon: <AlertCircle size={10} />, label: 'Failed' },
    [INSTANCE_STATUS.PROVISIONING]: { color: 'bg-blue-500/10 text-blue-500', icon: <Loader2 size={10} className="animate-spin" />, label: 'Provisioning' }
  }[status] || { color: 'bg-slate-800 text-slate-500', label: status };

  return (
    <div className={`flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[9px] font-black uppercase tracking-widest border border-current opacity-80 ${cfg.color}`}>
      {cfg.icon} {cfg.label}
    </div>
  );
};

const FormField = ({ label, value, onChange, placeholder, type = 'text', disabled = false }) => (
  <div className="space-y-1">
    <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">{label}</label>
    <input
      disabled={disabled}
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className="w-full bg-slate-900 border border-slate-800 rounded-xl px-4 py-2.5 text-xs outline-none focus:ring-1 ring-blue-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
    />
  </div>
);

const SectionHeader = ({ title, icon, action }) => (
  <div className="flex items-center justify-between mb-2">
    <div className="flex items-center gap-2">
      <span className="text-blue-500">{icon}</span>
      <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-[0.2em]">{title}</h3>
    </div>
    {action}
  </div>
);

const ApiErrorBanner = ({ message }) => (
  <div className="bg-red-500/10 border border-red-500/30 rounded-2xl p-4 flex items-center gap-3 text-red-500 text-xs font-bold">
    <AlertCircle size={18} />
    <span>Error de Sincronizacion: {message}</span>
  </div>
);

const LoadingState = () => (
  <div className="min-h-screen bg-[#020617] flex flex-col items-center justify-center gap-4">
    <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
    <span className="text-[10px] font-black uppercase text-slate-500 tracking-[0.3em] animate-pulse">Cargando Infraestructura...</span>
  </div>
);

const EmptyState = ({ message }) => (
  <div className="py-12 text-center text-slate-600 text-xs italic">{message}</div>
);

// --- FASE 2: MEJORAS ---

const SystemLogConsole = () => {
  const [logs] = useState([
    { time: '10:32:01', msg: 'System Audit Service started.', type: 'sys' },
    { time: '10:32:05', msg: 'Connected to Tailscale Control Plane.', type: 'net' },
    { time: '10:34:12', msg: 'Snapshotting product catalog...', type: 'db' }
  ]);

  return (
    <div className="bg-slate-900/20 border border-slate-800 rounded-2xl overflow-hidden h-40 flex flex-col">
      <div className="px-4 py-2 bg-slate-800/20 border-b border-slate-800 flex items-center gap-2">
        <Terminal size={12} className="text-slate-500" />
        <span className="text-[9px] font-black text-slate-500 uppercase tracking-widest">Auditoria del Sistema</span>
      </div>
      <div className="flex-1 p-4 font-mono text-[9px] space-y-1 overflow-y-auto custom-scrollbar bg-black/20">
        {logs.map((l, i) => (
          <div key={i} className="flex gap-4 opacity-60">
            <span className="text-slate-700 shrink-0">[{l.time}]</span>
            <span className="text-slate-400">{l.msg}</span>
          </div>
        ))}
      </div>
    </div>
  );
};

const CustomScrollbarStyles = () => (
  <style>{`
    .custom-scrollbar::-webkit-scrollbar { width: 4px; }
    .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
    .custom-scrollbar::-webkit-scrollbar-thumb { background: #1e293b; border-radius: 10px; }
    .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #334155; }
  `}</style>
);
