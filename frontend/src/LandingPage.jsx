import React, { useState, useEffect } from 'react';
import {
  Download, Shield, Cpu, Globe, CheckCircle,
  Terminal, Package, ExternalLink, X,
  Loader2, AlertCircle, Zap, ChevronDown, ChevronUp
} from 'lucide-react';

// --- CONFIGURACION Y CONSTANTES ---
const PRODUCT_THEMES = {
  media: { icon: Globe, color: 'text-blue-400', bg: 'bg-blue-500/10' },
  db: { icon: Cpu, color: 'text-purple-400', bg: 'bg-purple-500/10' },
  proxy: { icon: Shield, color: 'text-emerald-400', bg: 'bg-emerald-500/10' },
  default: { icon: Package, color: 'text-slate-400', bg: 'bg-slate-500/10' }
};

const dbg = (...args) => console.log('[Landing]', ...args);

// --- COMPONENTE PRINCIPAL ---
export default function ArkLanding() {
  // Estados de Datos
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [devices, setDevices] = useState([]);
  const [activeDeployment, setActiveDeployment] = useState(null);

  // Estados de UI
  const [isDeploying, setIsDeploying] = useState(false);
  const [error, setError] = useState(null);
  const [showLogs, setShowLogs] = useState(false);

  // 1. Carga Inicial y Recuperacion de Sesion
  useEffect(() => {
    const init = async () => {
      dbg('Init start');
      await Promise.all([fetchProducts(), fetchDevices(), recoverDeployment()]);
      dbg('Init done');
      setLoading(false);
    };
    init();
  }, []);

  const fetchProducts = async () => {
    try {
      dbg('GET /api/products');
      const res = await fetch('/api/products');
      dbg('GET /api/products status', res.status);
      if (!res.ok) throw new Error('No se pudo cargar el catalogo.');
      const data = await res.json();
      dbg('Products loaded', data.products?.length || 0);
      setProducts(data.products || []);
    } catch (err) {
      console.error('[Landing] fetchProducts error', err);
      setError('Error de conexion con el catalogo de productos.');
    }
  };

  const fetchDevices = async () => {
    try {
      dbg('GET /api/tailscale/devices');
      const res = await fetch('/api/tailscale/devices');
      dbg('GET /api/tailscale/devices status', res.status);
      if (res.ok) {
        const data = await res.json();
        dbg('Devices loaded', data.devices?.length || 0);
        setDevices(data.devices || []);
      }
    } catch (e) {
      console.error('[Landing] fetchDevices error', e);
    }
  };

  const recoverDeployment = async () => {
    try {
      dbg('GET /api/deployments (recover)');
      const res = await fetch('/api/deployments');
      dbg('GET /api/deployments status', res.status);
      if (res.ok) {
        const data = await res.json();
        const list = data.instances || [];
        dbg('Recovered instances', list.length);
        if (list.length > 0) {
          const last = list[0];
          dbg('Recovering active deployment', last.id, last.url, last.status);
          setActiveDeployment({
            instanceId: last.id,
            url: last.url,
            status: last.status || 'success',
            productName: last.product_id || 'Instancia Activa'
          });
        }
      }
    } catch (e) {
      console.log('[Landing] No se encontraron despliegues previos.');
    }
  };

  // 2. Logica de Despliegue
  const handleDeploy = async (product) => {
    if (isDeploying) return;
    dbg('Deploy click', product?.id, product?.name);

    setError(null);
    setIsDeploying(true);
    setActiveDeployment({ status: 'queued', productName: product.name });

    try {
      // Buscar host disponible
      const targetDevice = devices.find((d) => d.online || d.status === 'active');
      const targetHost = targetDevice?.addresses?.find((a) => a.startsWith('100.')) || null;

      if (!targetHost) {
        dbg('No target host from devices', devices);
        throw new Error('No hay nodos disponibles en la red para el despliegue.');
      }

      const payload = {
        product_id: product.id,
        environment: 'prod',
        target_host: targetHost,
        ssh_user: 'root'
      };
      dbg('POST /api/deployments payload', payload);

      const res = await fetch('/api/deployments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      dbg('POST /api/deployments status', res.status);

      const data = await res.json();
      dbg('POST /api/deployments response', data);

      if (!res.ok) {
        throw new Error(data.detail || 'Error interno del servidor al desplegar.');
      }

      setActiveDeployment({
        instanceId: data.instance_id,
        url: data.url,
        status: 'success',
        productName: product.name
      });
    } catch (err) {
      console.error('[Landing] handleDeploy error', err);
      setError(err.message);
      setActiveDeployment(null);
    } finally {
      dbg('Deploy finished');
      setIsDeploying(false);
    }
  };

  if (loading) return <LoadingSkeleton />;

  return (
    <div className="min-h-screen bg-[#020617] text-slate-200 selection:bg-blue-500/30">
      <Navbar />

      <main className="max-w-4xl mx-auto px-6 pt-32 pb-20">
        {error && <ErrorBanner message={error} onClose={() => setError(null)} />}

        <div className="text-center mb-16">
          <h1 className="text-4xl md:text-6xl font-black text-white mb-6 tracking-tighter">
            Despliegue <span className="text-blue-500">Ark</span>
          </h1>
          <p className="text-slate-400 font-medium mb-12">
            Infraestructura privada lista para produccion en un clic.
          </p>

          <div className="max-w-md mx-auto">
            {products.length > 0 ? (
              <div className="bg-slate-900 border border-slate-800 p-8 rounded-[2.5rem] shadow-2xl shadow-black/50 animate-in fade-in zoom-in-95 duration-500">
                <div className="flex flex-col items-center gap-6">
                  <ProductIcon productId={products[0].id} />

                  <div className="space-y-1">
                    <h2 className="text-2xl font-bold text-white tracking-tight">
                      {products[0].name}
                    </h2>
                    <p className="text-[10px] font-mono text-slate-500 uppercase tracking-[0.2em]">Version Estable</p>
                  </div>

                  <button
                    onClick={() => handleDeploy(products[0])}
                    disabled={isDeploying}
                    className={`w-full py-5 rounded-2xl font-black uppercase tracking-widest text-sm transition-all flex items-center justify-center gap-3 ${
                      isDeploying
                        ? 'bg-slate-800 text-slate-500 cursor-not-allowed'
                        : 'bg-blue-600 hover:bg-blue-500 text-white shadow-xl shadow-blue-500/20 active:scale-95'
                    }`}
                  >
                    {isDeploying ? <Loader2 className="animate-spin" size={18} /> : <Download size={18} />}
                    {isDeploying ? 'Desplegando...' : 'Desplegar'}
                  </button>
                </div>
              </div>
            ) : (
              <div className="p-8 border border-dashed border-slate-800 rounded-3xl text-slate-500 text-xs uppercase tracking-widest font-bold">
                No hay productos disponibles
              </div>
            )}
          </div>
        </div>

        {activeDeployment && (
          <div className="space-y-6 animate-fade-in">
            <DeployStatusCard deployment={activeDeployment} />

            {activeDeployment.status === 'success' && (
              <InstanceAccessCard data={activeDeployment} />
            )}

            <LogsPanel
              isOpen={showLogs}
              onToggle={() => setShowLogs(!showLogs)}
              status={activeDeployment.status}
            />
          </div>
        )}
      </main>

      <SupportFooter />
    </div>
  );
}

// --- SUB-COMPONENTES ---

const ProductIcon = ({ productId }) => {
  const themeKey = Object.keys(PRODUCT_THEMES).find((k) => productId.toLowerCase().includes(k)) || 'default';
  const theme = PRODUCT_THEMES[themeKey];
  const Icon = theme.icon;
  return (
    <div className={`w-16 h-16 ${theme.bg} rounded-3xl flex items-center justify-center border border-white/5`}>
      <Icon className={theme.color} size={32} />
    </div>
  );
};

const Navbar = () => (
  <nav className="fixed top-0 w-full z-50 bg-[#020617]/80 backdrop-blur-md border-b border-white/5">
    <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
      <div className="flex items-center gap-2">
        <Zap size={20} className="text-blue-500 fill-current" />
        <span className="font-black text-lg tracking-tighter uppercase">Ark</span>
      </div>
      <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
        Sistema de Orquestacion
      </div>
    </div>
  </nav>
);

const ErrorBanner = ({ message, onClose }) => (
  <div className="mb-8 bg-red-500/10 border border-red-500/20 p-4 rounded-xl flex items-center justify-between text-red-400 animate-in slide-in-from-top-4">
    <div className="flex items-center gap-3 text-sm font-medium">
      <AlertCircle size={18} />
      {message}
    </div>
    <button onClick={onClose} className="hover:text-white"><X size={18} /></button>
  </div>
);

const DeployStatusCard = ({ deployment }) => {
  const isRunning = deployment.status === 'queued' || deployment.status === 'running';
  return (
    <div className="bg-slate-900 border border-slate-800 p-6 rounded-3xl flex items-center justify-between">
      <div className="flex items-center gap-4">
        <div className={`w-12 h-12 rounded-2xl flex items-center justify-center ${isRunning ? 'bg-blue-500/10' : 'bg-emerald-500/10'}`}>
          {isRunning ? <Loader2 className="text-blue-500 animate-spin" /> : <CheckCircle className="text-emerald-500" />}
        </div>
        <div>
          <h4 className="font-bold text-white">{deployment.productName}</h4>
          <p className="text-xs text-slate-500 uppercase tracking-widest font-mono">
            Estado: {deployment.status === 'success' ? 'Ejecutando' : 'En cola'}
          </p>
        </div>
      </div>
      <div className="hidden md:block">
        <span className="text-[10px] font-mono text-slate-600 italic">ID: {deployment.instanceId || 'PENDIENTE'}</span>
      </div>
    </div>
  );
};

const InstanceAccessCard = ({ data }) => (
  <div className="bg-emerald-500 border border-emerald-400 p-1 rounded-[2rem] shadow-2xl shadow-emerald-500/20 animate-in zoom-in-95">
    <div className="bg-slate-950 rounded-[1.8rem] p-8">
      <h4 className="text-emerald-400 font-black uppercase text-xs tracking-widest mb-4">Instancia Lista</h4>
      <div className="flex flex-col md:flex-row gap-4 items-center">
        <div className="flex-1 w-full bg-slate-900 p-4 rounded-2xl border border-slate-800">
          <p className="text-[10px] text-slate-500 uppercase font-bold mb-1">Direccion de Acceso</p>
          <p className="font-mono text-blue-400 truncate">{data.url}</p>
        </div>
        <a
          href={data.url}
          target="_blank"
          rel="noreferrer"
          className="w-full md:w-auto px-8 py-4 bg-emerald-500 hover:bg-emerald-400 text-slate-950 font-black uppercase text-xs tracking-widest rounded-2xl transition-all flex items-center justify-center gap-2"
        >
          Abrir <ExternalLink size={16} />
        </a>
      </div>
    </div>
  </div>
);

const LogsPanel = ({ isOpen, onToggle, status }) => (
  <div className="border border-slate-800 rounded-2xl overflow-hidden">
    <button
      onClick={onToggle}
      className="w-full px-6 py-4 bg-slate-900/50 flex items-center justify-between hover:bg-slate-900 transition-colors"
    >
      <div className="flex items-center gap-2 text-xs font-bold text-slate-400 uppercase tracking-widest">
        <Terminal size={14} /> Registros
      </div>
      {isOpen ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
    </button>
    {isOpen && (
      <div className="bg-black p-6 font-mono text-[10px] text-slate-400 h-48 overflow-y-auto space-y-1">
        <p className="text-blue-500 opacity-50">[SYSTEM] Iniciando monitor...</p>
        <p>[INFO] Verificando SSH root...</p>
        {status === 'success' && <p className="text-emerald-500">[OK] Proceso completado.</p>}
      </div>
    )}
  </div>
);

const LoadingSkeleton = () => (
  <div className="min-h-screen bg-[#020617] flex flex-col items-center justify-center gap-4">
    <div className="w-10 h-10 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
    <span className="text-[10px] font-black text-slate-500 uppercase tracking-[0.3em]">Ark</span>
  </div>
);

const SupportFooter = () => (
  <footer className="max-w-4xl mx-auto px-6 py-10 border-t border-slate-900 mt-20">
    <div className="flex flex-col md:flex-row justify-between items-center gap-6">
      <p className="text-slate-600 text-[10px] uppercase font-bold tracking-widest">
        © 2026 Ark Cloud
      </p>
      <div className="flex gap-8 text-[10px] font-black uppercase tracking-widest text-slate-500">
        <a href="mailto:soporte@ark.io" className="hover:text-blue-500 transition-colors">Soporte</a>
        <a href="#" className="hover:text-blue-500 transition-colors">Docs</a>
      </div>
    </div>
  </footer>
);

if (typeof document !== 'undefined') {
  const styleTagId = 'ark-landing-animations';
  if (!document.getElementById(styleTagId)) {
    const style = document.createElement('style');
    style.id = styleTagId;
    style.textContent = `
      @keyframes fade-in { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
      .animate-fade-in { animation: fade-in 0.8s ease-out forwards; }
    `;
    document.head.appendChild(style);
  }
}
