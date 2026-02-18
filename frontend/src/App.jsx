import React, { useState, useEffect, useRef } from 'react';
import { 
  Play, 
  RefreshCw, 
  Trash2, 
  Activity, 
  CheckCircle2, 
  Clock, 
  AlertCircle, 
  Server, 
  Terminal,
  Search,
  Settings,
  ShieldCheck,
  Globe,
  Link2,
  Cpu,
  Copy,
  Eraser,
  ExternalLink,
  Package,
  Tag,
  Layers,
  X,
  Plus,
  ArrowRight,
  ShieldAlert,
  GitBranch,
  Network,
  Share2
} from 'lucide-react';

const INITIAL_PRODUCTS = [
  { 
    id: 'prod-1', 
    name: 'ARK Survival Evolved', 
    version: 'v2.4.1', 
    env: 'PROD',
    status: 'success', 
    lastRun: 'Hace 2 horas', 
    buildNum: '#42', 
    description: 'Instancia principal del servidor de juego con persistencia activa.'
  },
  { 
    id: 'prod-2', 
    name: 'ARK Cluster Maps', 
    version: 'v1.0.5', 
    env: 'STAGING',
    status: 'idle', 
    lastRun: 'Ayer', 
    buildNum: '#12', 
    description: 'Sincronización de mapas y transferencias de personajes entre servidores.'
  }
];

const INITIAL_NODES = [
  { name: 'ark-prod-01', ip: '100.82.15.42', status: 'active', region: 'Local', type: 'Exit Node' },
  { name: 'ark-backup-02', ip: '100.10.5.20', status: 'idle', region: 'Cloud', type: 'Relay' },
  { name: 'dev-node-razie', ip: '100.12.0.5', status: 'offline', region: 'Local', type: 'Direct' },
  { name: 'storage-vault', ip: '100.40.1.15', status: 'active', region: 'Edge', type: 'Direct' }
];

export default function App() {
  const [products, setProducts] = useState(INITIAL_PRODUCTS);
  const [tailscaleNodes, setTailscaleNodes] = useState(INITIAL_NODES);
  const [filter, setFilter] = useState('');
  const [activeExecution, setActiveExecution] = useState(null);
  const [followLogs, setFollowLogs] = useState(true);
  
  const [isDeployModalOpen, setIsDeployModalOpen] = useState(false);
  const [isNewDeviceModalOpen, setIsNewDeviceModalOpen] = useState(false);
  const [selectedProductForDeploy, setSelectedProductForDeploy] = useState(null);
  const [tempTargetHost, setTempTargetHost] = useState('');

  const [logs, setLogs] = useState([
    { id: 1, time: '14:30:05', msg: '[System] ARK_DEPLOY backend ready.', type: 'sys' },
    { id: 2, time: '14:31:10', msg: '[Tailscale] Connected to private mesh.', type: 'sys' }
  ]);

  const logEndRef = useRef(null);

  useEffect(() => {
    if (followLogs && logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs, followLogs]);

  // Lógica para el comando TREE visual
  const groupedNodes = tailscaleNodes.reduce((acc, node) => {
    if (!acc[node.region]) acc[node.region] = [];
    acc[node.region].push(node);
    return acc;
  }, {});

  const openDeployModal = (product) => {
    setSelectedProductForDeploy(product);
    setTempTargetHost(tailscaleNodes[0]?.ip || '');
    setIsDeployModalOpen(true);
  };

  const confirmDeployment = () => {
    const host = tailscaleNodes.find(n => n.ip === tempTargetHost);
    if (!host || host.status === 'offline') return;

    setIsDeployModalOpen(false);
    setActiveExecution({
      productName: selectedProductForDeploy.name,
      target: host.name,
      status: 'provisioning',
      buildId: `#${Math.floor(Math.random() * 100) + 100}`
    });

    addLog(`Desplegando ${selectedProductForDeploy.name} en ${host.name}...`, 'info');
    setTimeout(() => {
      setActiveExecution(prev => prev ? ({ ...prev, status: 'running' }) : null);
      addLog(`[Jenkins] Pipeline iniciado en nodo remoto.`, 'info');
    }, 1500);
  };

  const addNewDevice = (e) => {
    e.preventDefault();
    const formData = new FormData(e.target);
    const newNode = {
      name: formData.get('name'),
      ip: `100.${Math.floor(Math.random()*255)}.${Math.floor(Math.random()*255)}.${Math.floor(Math.random()*255)}`,
      status: 'active',
      region: formData.get('region'),
      type: 'Direct'
    };
    setTailscaleNodes([...tailscaleNodes, newNode]);
    addLog(`[Tailscale] Dispositivo '${newNode.name}' vinculado a la red.`, 'sys');
    setIsNewDeviceModalOpen(false);
  };

  const addLog = (msg, type = 'info') => {
    setLogs(prev => [...prev, { 
      id: Date.now() + Math.random(), 
      time: new Date().toLocaleTimeString(), 
      msg, 
      type 
    }]);
  };

  const getLogColor = (type) => {
    switch(type) {
      case 'err': return 'text-red-400';
      case 'stage': return 'text-cyan-400 font-bold';
      case 'sys': return 'text-slate-500 italic';
      default: return 'text-slate-300';
    }
  };

  return (
    <div className="min-h-screen bg-[#020617] text-slate-100 font-sans p-4 md:p-8 flex flex-col selection:bg-blue-500/30">
      
      {/* Header */}
      <header className="max-w-7xl w-full mx-auto mb-8 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="p-3 bg-blue-600/10 rounded-2xl border border-blue-500/20 shadow-[0_0_20px_rgba(59,130,246,0.1)]">
            <Package className="text-blue-400 w-8 h-8" />
          </div>
          <div>
            <h1 className="text-2xl font-black tracking-tighter uppercase">ARK <span className="text-blue-500">Products</span></h1>
            <div className="flex items-center gap-2 text-[10px] text-slate-500 font-bold uppercase tracking-widest">
              <span className="flex items-center gap-1 text-green-500"><ShieldCheck size={12}/> VPN: {tailscaleNodes.length} Nodes</span>
              <span className="text-slate-800">|</span>
              <span className="flex items-center gap-1"><Network size={12}/> Mesh Topology Active</span>
            </div>
          </div>
        </div>

        <button 
          onClick={() => setIsNewDeviceModalOpen(true)}
          className="flex items-center gap-2 px-5 py-2.5 bg-slate-900 border border-slate-800 rounded-xl text-[11px] font-black uppercase hover:bg-slate-800 transition-all text-slate-300 tracking-wider shadow-lg"
        >
          <Plus size={16} className="text-blue-400" /> New Device
        </button>
      </header>

      <main className="max-w-7xl w-full mx-auto grid grid-cols-1 lg:grid-cols-12 gap-6 flex-1">
        
        {/* Lado Izquierdo: Productos y Tree View */}
        <div className="lg:col-span-8 space-y-6">
          
          {/* Active Task (Si hay una) */}
          {activeExecution && (
            <div className="bg-blue-600/5 border border-blue-500/30 rounded-2xl p-6 ring-1 ring-blue-500/10 relative overflow-hidden group">
              <div className="flex items-center justify-between relative z-10">
                <div className="flex items-center gap-5">
                  <div className="w-12 h-12 bg-blue-600 rounded-2xl flex items-center justify-center shadow-[0_0_20px_rgba(37,99,235,0.4)]">
                    <RefreshCw className="text-white animate-spin" size={24} />
                  </div>
                  <div>
                    <h2 className="text-lg font-bold tracking-tight">Deploying to {activeExecution.target}</h2>
                    <p className="text-sm text-blue-400/70 font-medium tracking-wide italic">{activeExecution.productName} • Build {activeExecution.buildId}</p>
                  </div>
                </div>
                <div className="bg-blue-500/20 px-4 py-1.5 rounded-full border border-blue-500/30">
                   <span className="text-[10px] text-blue-300 font-black uppercase tracking-widest animate-pulse">
                    {activeExecution.status}...
                  </span>
                </div>
              </div>
              <div className="absolute top-0 right-0 w-32 h-full bg-gradient-to-l from-blue-500/5 to-transparent group-hover:from-blue-500/10 transition-all duration-700" />
            </div>
          )}

          {/* Catálogo de Productos */}
          <section className="space-y-4">
             <div className="flex items-center justify-between px-2">
              <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-[0.3em] flex items-center gap-2">
                <Layers size={14} className="text-blue-500" /> Deployment Units
              </h3>
              <div className="relative">
                <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600" />
                <input 
                  type="text" 
                  placeholder="Filter by name..."
                  className="bg-slate-900/50 border border-slate-800 rounded-xl pl-9 pr-4 py-1.5 text-xs focus:ring-1 ring-blue-500 outline-none w-48 transition-all focus:w-64 placeholder:text-slate-700 text-white"
                  onChange={(e) => setFilter(e.target.value)}
                />
              </div>
            </div>

            <div className="grid gap-3">
              {products.filter(p => p.name.toLowerCase().includes(filter.toLowerCase())).map((product) => (
                <div 
                  key={product.id} 
                  className="bg-slate-900/40 border border-slate-800/80 rounded-2xl p-6 hover:border-blue-500/30 transition-all group relative overflow-hidden"
                >
                  <div className="flex flex-col lg:flex-row justify-between gap-6 relative z-10">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h4 className="text-lg font-bold text-slate-100 group-hover:text-blue-400 transition-colors">{product.name}</h4>
                        <span className={`text-[9px] px-2 py-0.5 rounded-md font-black tracking-widest ${product.env === 'PROD' ? 'bg-red-500/10 text-red-500 border border-red-500/20' : 'bg-yellow-500/10 text-yellow-500 border border-yellow-500/20'}`}>
                          {product.env}
                        </span>
                        <div className="flex items-center gap-1.5 text-[10px] text-slate-600 font-mono">
                          <Tag size={10} /> {product.version}
                        </div>
                      </div>
                      <p className="text-sm text-slate-500 max-w-2xl leading-relaxed font-medium">
                        {product.description}
                      </p>
                      
                      <div className="flex items-center gap-5 mt-5 text-[10px] font-black text-slate-600 uppercase tracking-widest">
                        <span className="flex items-center gap-2 border-r border-slate-800 pr-4"><CheckCircle2 size={13} className={product.status === 'success' ? 'text-green-500' : 'text-slate-700'}/> Last Artifact: {product.buildNum}</span>
                        <span className="flex items-center gap-2"><Clock size={13}/> Updated: {product.lastRun}</span>
                      </div>
                    </div>

                    <div className="flex items-center shrink-0">
                      <button 
                        onClick={() => openDeployModal(product)}
                        className="flex items-center justify-center gap-3 px-8 py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-black uppercase transition-all shadow-[0_10px_30px_rgba(37,99,235,0.2)] active:scale-95 group-hover:translate-x-1"
                      >
                        <Play size={14} fill="white" /> New Deployment
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* TAILSCALE TREE VIEW (Nueva sección solicitada) */}
          <section className="bg-slate-900/20 border border-slate-800/60 rounded-3xl p-6">
            <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-[0.3em] mb-6 flex items-center gap-2">
              <GitBranch size={14} className="text-blue-500" /> Tailscale Network Tree
            </h3>
            
            <div className="font-mono text-[11px] space-y-4">
              {/* Root Node */}
              <div className="flex items-center gap-2 text-blue-400 font-bold">
                <Globe size={14} />
                <span>tailnet.ark-servers.mesh</span>
              </div>

              {Object.keys(groupedNodes).map((region, regionIdx, regionArr) => (
                <div key={region} className="ml-4 border-l border-slate-800 pl-4 relative">
                  {/* Conector del Árbol */}
                  <div className="absolute left-0 top-3 w-4 h-px bg-slate-800" />
                  
                  <div className="flex items-center gap-2 mb-3 text-slate-400 uppercase font-black tracking-widest text-[9px]">
                    <span className="text-slate-700">├──</span>
                    <Server size={12} className="text-slate-600" />
                    <span>Region: {region}</span>
                  </div>

                  <div className="space-y-3 ml-4">
                    {groupedNodes[region].map((node, nodeIdx, nodeArr) => (
                      <div key={node.ip} className="flex items-start gap-3 relative group">
                        <span className="text-slate-700 font-bold shrink-0">
                          {nodeIdx === nodeArr.length - 1 ? '└──' : '├──'}
                        </span>
                        
                        <div className={`p-3 rounded-xl border border-slate-800/60 bg-slate-950/40 flex-1 flex items-center justify-between group-hover:border-slate-700 transition-colors ${node.status === 'offline' ? 'opacity-40 grayscale' : ''}`}>
                          <div className="flex items-center gap-4">
                             <div className={`w-2.5 h-2.5 rounded-full ${node.status === 'active' ? 'bg-green-500 shadow-[0_0_10px_rgba(34,197,94,0.4)]' : node.status === 'idle' ? 'bg-yellow-500' : 'bg-slate-800'}`} />
                             <div>
                                <p className="font-bold text-slate-200 tracking-tight group-hover:text-blue-400 transition-colors uppercase">{node.name}</p>
                                <p className="text-[9px] text-slate-600 font-mono tracking-tighter">{node.ip} · {node.type}</p>
                             </div>
                          </div>
                          
                          <div className="flex items-center gap-4">
                            <div className="text-right">
                               <p className="text-[8px] text-slate-700 font-black uppercase mb-1">Latency</p>
                               <div className="flex gap-0.5">
                                 {[1,2,3,4].map(b => (
                                   <div key={b} className={`w-1 h-2 rounded-full ${node.status === 'active' && b <= 3 ? 'bg-blue-500' : 'bg-slate-800'}`} />
                                 ))}
                               </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </section>
        </div>

        {/* Lado Derecho: Live Console */}
        <div className="lg:col-span-4 flex flex-col gap-6">
          <section className="bg-black border border-slate-800/80 rounded-[2rem] overflow-hidden flex flex-col h-full min-h-[600px] shadow-2xl">
            <div className="bg-[#0f172a] px-6 py-4 border-b border-slate-800/60 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Terminal size={16} className="text-blue-500" />
                <span className="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">Deployment Stream</span>
              </div>
              <button 
                onClick={() => setLogs([])}
                className="p-1.5 hover:bg-slate-800 rounded-lg text-slate-600 hover:text-red-400 transition-all"
              >
                <Eraser size={16} />
              </button>
            </div>
            
            <div className="p-6 font-mono text-[11px] overflow-y-auto flex-1 space-y-3 bg-[#020617] custom-scrollbar">
              {logs.map((log) => (
                <div key={log.id} className="flex gap-4 leading-normal group">
                  <span className="text-slate-800 shrink-0 select-none font-bold">[{log.time}]</span>
                  <span className={`${getLogColor(log.type)} tracking-tight break-all`}>
                    {log.type === 'stage' && '>> '} {log.msg}
                  </span>
                </div>
              ))}
              <div ref={logEndRef} />
            </div>

            <div className="p-4 border-t border-slate-900 bg-[#0f172a]/40 flex justify-between items-center text-[10px] font-black text-slate-600 uppercase tracking-widest">
               <div className="flex items-center gap-2">
                 <div className="w-1.5 h-1.5 rounded-full bg-blue-500 animate-ping" />
                 <span>Socket Connected</span>
               </div>
               <Share2 size={14} className="opacity-40" />
            </div>
          </section>
        </div>
      </main>

      {/* MODAL: Despliegue */}
      {isDeployModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-[#020617]/90 backdrop-blur-md">
          <div className="bg-slate-950 border border-slate-800 w-full max-w-lg rounded-[2rem] shadow-[0_0_100px_rgba(59,130,246,0.15)] overflow-hidden">
            <div className="px-8 py-6 border-b border-slate-800 flex items-center justify-between">
              <h2 className="font-black uppercase tracking-[0.2em] text-xs flex items-center gap-3">
                <Play size={18} className="text-blue-500" /> Confirm Deployment
              </h2>
              <button onClick={() => setIsDeployModalOpen(false)} className="text-slate-500 hover:text-white transition-colors">
                <X size={24} />
              </button>
            </div>
            <div className="p-8 space-y-8">
              <div className="space-y-3">
                <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">Unit to Deploy</label>
                <div className="bg-slate-900/50 p-5 rounded-2xl border border-slate-800/80 flex items-center gap-4">
                   <div className="p-3 bg-blue-500/10 rounded-xl"><Package className="text-blue-400" size={24}/></div>
                   <div>
                     <p className="text-lg font-bold text-slate-100">{selectedProductForDeploy?.name}</p>
                     <p className="text-xs text-blue-400/70 font-mono">{selectedProductForDeploy?.version} · {selectedProductForDeploy?.env}</p>
                   </div>
                </div>
              </div>

              <div className="space-y-3">
                <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">Select Target Endpoint</label>
                <div className="grid gap-2 max-h-60 overflow-y-auto custom-scrollbar pr-2">
                  {tailscaleNodes.map(node => (
                    <button
                      key={node.ip}
                      disabled={node.status === 'offline'}
                      onClick={() => setTempTargetHost(node.ip)}
                      className={`w-full p-4 rounded-2xl border text-left transition-all flex items-center justify-between ${
                        tempTargetHost === node.ip 
                        ? 'bg-blue-600 border-blue-500 shadow-lg shadow-blue-600/20' 
                        : 'bg-slate-900/50 border-slate-800 hover:border-slate-700'
                      } ${node.status === 'offline' ? 'opacity-20 cursor-not-allowed' : ''}`}
                    >
                      <div className="flex items-center gap-4">
                        <div className={`w-2 h-2 rounded-full ${tempTargetHost === node.ip ? 'bg-white shadow-[0_0_10px_#fff]' : 'bg-green-500'}`} />
                        <div>
                          <p className={`text-xs font-black uppercase ${tempTargetHost === node.ip ? 'text-white' : 'text-slate-300'}`}>{node.name}</p>
                          <p className={`text-[10px] font-mono ${tempTargetHost === node.ip ? 'text-blue-100' : 'text-slate-600'}`}>{node.ip}</p>
                        </div>
                      </div>
                      {tempTargetHost === node.ip && <CheckCircle2 size={18} className="text-white" />}
                    </button>
                  ))}
                </div>
              </div>
            </div>
            <div className="p-8 bg-black/40 border-t border-slate-800 flex gap-4">
              <button 
                onClick={() => setIsDeployModalOpen(false)}
                className="flex-1 px-6 py-4 bg-slate-900 hover:bg-slate-800 rounded-2xl text-[11px] font-black uppercase transition-all tracking-widest text-slate-500"
              >
                Cancel
              </button>
              <button 
                onClick={confirmDeployment}
                className="flex-[2] px-6 py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl text-[11px] font-black uppercase transition-all shadow-[0_15px_40px_rgba(37,99,235,0.3)] flex items-center justify-center gap-3 tracking-[0.2em]"
              >
                Execute Pipeline <ArrowRight size={18} />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL: Nuevo Dispositivo */}
      {isNewDeviceModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-[#020617]/90 backdrop-blur-md">
          <div className="bg-slate-950 border border-slate-800 w-full max-w-md rounded-[2rem] shadow-2xl overflow-hidden">
            <div className="px-8 py-6 border-b border-slate-800 flex items-center justify-between">
              <h2 className="font-black uppercase tracking-[0.2em] text-xs flex items-center gap-3 text-slate-400">
                <Link2 size={18} className="text-blue-500" /> Add Node to Mesh
              </h2>
              <button onClick={() => setIsNewDeviceModalOpen(false)} className="text-slate-500 hover:text-white transition-colors">
                <X size={24} />
              </button>
            </div>
            <form onSubmit={addNewDevice} className="p-8 space-y-6">
              <div className="bg-blue-600/5 p-4 rounded-2xl border border-blue-500/20 flex gap-4">
                <ShieldAlert className="text-blue-500 shrink-0" size={20} />
                <p className="text-[11px] text-slate-500 leading-relaxed font-bold uppercase tracking-tight">
                  Dispositivo debe tener Tailscale pre-instalado para el handshake inicial.
                </p>
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block ml-1">Device Hostname</label>
                <input 
                  required
                  name="name"
                  type="text" 
                  placeholder="ej. ark-node-tx-01"
                  className="w-full bg-slate-900 border border-slate-800 rounded-2xl px-5 py-4 text-sm focus:ring-1 ring-blue-500 outline-none text-white placeholder:text-slate-700"
                />
              </div>
              <div className="space-y-2">
                <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block ml-1">Deployment Zone</label>
                <select 
                  name="region"
                  className="w-full bg-slate-900 border border-slate-800 rounded-2xl px-5 py-4 text-sm outline-none cursor-pointer text-white appearance-none"
                >
                  <option value="Local">Local Data Center</option>
                  <option value="Cloud">Public Cloud (AWS)</option>
                  <option value="Edge">Edge Node</option>
                </select>
              </div>
              <button 
                type="submit"
                className="w-full mt-4 py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl text-[11px] font-black uppercase tracking-[0.2em] shadow-[0_10px_30px_rgba(37,99,235,0.2)] transition-all active:scale-95"
              >
                Authorize Connection
              </button>
            </form>
          </div>
        </div>
      )}

      <style>{`
        .custom-scrollbar::-webkit-scrollbar {
          width: 5px;
        }
        .custom-scrollbar::-webkit-scrollbar-track {
          background: transparent;
        }
        .custom-scrollbar::-webkit-scrollbar-thumb {
          background: #1e293b;
          border-radius: 10px;
        }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover {
          background: #3b82f6;
        }
      `}</style>
    </div>
  );
}
