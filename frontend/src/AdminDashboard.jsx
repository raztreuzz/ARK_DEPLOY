import React, { useState, useEffect, useRef } from 'react';
import { 
  Play, RefreshCw, Activity, CheckCircle2, Clock, 
  Server, Terminal, Search, ShieldCheck, Globe, 
  Cpu, Copy, Eraser, ExternalLink, Package, Tag, 
  Layers, X, ArrowRight, ShieldAlert, GitBranch, 
  Network, Share2, Link2, Trash2, AlertCircle, FileText, Loader
} from 'lucide-react';

// Constants
const LOG_TYPES = {
  ERROR: 'err',
  STAGE: 'stage',
  SYSTEM: 'sys',
  INFO: 'info'
};

const LOG_COLORS = {
  [LOG_TYPES.ERROR]: 'text-red-400',
  [LOG_TYPES.STAGE]: 'text-cyan-400 font-bold',
  [LOG_TYPES.SYSTEM]: 'text-slate-500 italic',
  [LOG_TYPES.INFO]: 'text-slate-300'
};

const DEPLOYMENT_STATUS = {
  PROVISIONING: 'provisioning',
  RUNNING: 'running',
  SUCCESS: 'success',
  FAILED: 'failed'
};

const DEFAULT_PRODUCT_VALUES = {
  version: 'v1.0.0',
  env: 'PROD',
  status: 'idle',
  lastRun: 'Never',
  buildNum: '-'
};

const DEFAULT_NODE_VALUES = {
  status: 'idle',
  region: 'Cloud',
  type: 'Direct'
};

export default function AdminDashboard() {
  const [products, setProducts] = useState([]);
  const [tailscaleNodes, setTailscaleNodes] = useState([]);
  const [instances, setInstances] = useState([]);
  const [filter, setFilter] = useState('');
  const [activeExecution, setActiveExecution] = useState(null);
  const [followLogs, setFollowLogs] = useState(true);
  const [isDeployModalOpen, setIsDeployModalOpen] = useState(false);

  const [selectedProductForDeploy, setSelectedProductForDeploy] = useState(null);
  const [tempTargetHost, setTempTargetHost] = useState('');
  const [logs, setLogs] = useState([]);
  const [deleteConfirmation, setDeleteConfirmation] = useState(null);
  const [logsModal, setLogsModal] = useState(null);
  const [isCreateProductOpen, setIsCreateProductOpen] = useState(false);
  const [productFormError, setProductFormError] = useState('');
  const [jobsCatalog, setJobsCatalog] = useState([]);
  const [jobsLoading, setJobsLoading] = useState(false);
  const [jobsError, setJobsError] = useState('');
  const [productForm, setProductForm] = useState({
    id: '',
    name: '',
    description: '',
    deployJobs: {
      PROD: '',
      DEV: ''
    },
    deleteJob: ''
  });

  const logEndRef = useRef(null);

  useEffect(() => {
    loadProducts();
    loadTailscaleDevices();
    loadInstances();
    addLog('[System] ARK_DEPLOY backend ready.', LOG_TYPES.SYSTEM);
    addLog('[Tailscale] Connecting to private mesh...', LOG_TYPES.SYSTEM);
  }, []);

  useEffect(() => {
    if (followLogs && logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [logs, followLogs]);

  const loadProducts = async () => {
    try {
      const response = await fetch('/api/products');
      if (!response.ok) {
        const detail = await response.json().catch(() => ({}));
        throw new Error(detail.detail || response.statusText);
      }
      const data = await response.json();
      if (data.products) {
        const mappedProducts = data.products.map(p => ({
          ...DEFAULT_PRODUCT_VALUES,
          id: p.id,
          name: p.name,
          description: p.description || 'No description available',
          deploy_jobs: p.deploy_jobs || p.jobs || {},
          delete_job: p.delete_job || ''
        }));
        setProducts(mappedProducts);
        addLog(`[API] Loaded ${mappedProducts.length} product(s) from backend.`, LOG_TYPES.INFO);
      }
    } catch (error) {
      console.error('Error loading products:', error);
      addLog(`[API] Failed to load products: ${error.message}`, LOG_TYPES.ERROR);
    }
  };

  const loadTailscaleDevices = async () => {
    try {
      const response = await fetch('/api/tailscale/devices');
      if (!response.ok) {
        const detail = await response.json().catch(() => ({}));
        throw new Error(detail.detail || detail.error || response.statusText);
      }
      const data = await response.json();
      if (data.devices) {
        const mappedDevices = data.devices.map(d => ({
          ...DEFAULT_NODE_VALUES,
          name: d.hostname || d.name,
          ip: d.addresses?.[0] || 'N/A',
          status: d.online ? 'active' : 'offline'
        }));
        setTailscaleNodes(mappedDevices);
        addLog(`[Tailscale] Connected. ${mappedDevices.length} node(s) in mesh.`, LOG_TYPES.SYSTEM);
      }
    } catch (error) {
      console.error('Error loading Tailscale devices:', error);
      addLog(`[Tailscale] Connection failed: ${error.message}`, LOG_TYPES.ERROR);
    }
  };

  const loadInstances = async () => {
    try {
      const response = await fetch('/api/deployments');
      if (!response.ok) {
        const detail = await response.json().catch(() => ({}));
        throw new Error(detail.detail || response.statusText);
      }
      const data = await response.json();
      if (data.instances) {
        setInstances(data.instances);
        addLog(`[Instances] Loaded ${data.instances.length} active instance(s).`, LOG_TYPES.INFO);
      }
    } catch (error) {
      console.error('Error loading instances:', error);
      addLog(`[Instances] Failed to load instances: ${error.message}`, LOG_TYPES.ERROR);
    }
  };

  const loadJobsCatalog = async () => {
    setJobsLoading(true);
    setJobsError('');
    try {
      const response = await fetch('/api/jobs');
      if (!response.ok) {
        const detail = await response.json().catch(() => ({}));
        throw new Error(detail.detail || response.statusText);
      }
      const data = await response.json();
      setJobsCatalog(data.jobs || []);
    } catch (error) {
      console.error('Error loading jobs:', error);
      setJobsCatalog([]);
      setJobsError(error.message || 'Failed to load jobs');
    } finally {
      setJobsLoading(false);
    }
  };

  const openDeployModal = (product) => {
    setSelectedProductForDeploy(product);
    setTempTargetHost(tailscaleNodes[0]?.ip || '');
    setIsDeployModalOpen(true);
  };

  const confirmDeployment = async () => {
    const host = tailscaleNodes.find(n => n.ip === tempTargetHost);
    if (!host || host.status === 'offline') return;

    setIsDeployModalOpen(false);
    setActiveExecution({
      productName: selectedProductForDeploy.name,
      target: host.name,
      status: DEPLOYMENT_STATUS.PROVISIONING,
      buildId: `#${Math.floor(Math.random() * 100) + 100}`
    });

    addLog(`Desplegando ${selectedProductForDeploy.name} en ${host.name}...`, LOG_TYPES.INFO);
    
    try {
      const response = await fetch('/api/deployments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          product_id: selectedProductForDeploy.id,
          target_host: host.ip,
          environment: selectedProductForDeploy.env
        })
      });

      if (response.ok) {
        const data = await response.json();
        setActiveExecution(prev => prev ? ({ 
          ...prev, 
          status: DEPLOYMENT_STATUS.RUNNING, 
          buildId: data.build_id || prev.buildId 
        }) : null);
        addLog(`[Jenkins] Pipeline iniciado: ${data.job_name || 'unknown'}`, LOG_TYPES.INFO);
        addLog(`[Build] Build ID: ${data.build_id || 'N/A'}`, LOG_TYPES.STAGE);
      } else {
        addLog(`[Error] Deployment failed: ${response.statusText}`, LOG_TYPES.ERROR);
        setActiveExecution(null);
      }
    } catch (error) {
      console.error('Error deploying:', error);
      addLog(`[Error] Network error during deployment.`, LOG_TYPES.ERROR);
      setActiveExecution(null);
    }
  };

  const handleDeleteInstance = async (instanceId, deviceId, productId) => {
    try {
      const response = await fetch(`/api/deployments/${instanceId}`, {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' }
      });

      if (response.ok) {
        setInstances(prev => prev.filter(i => i.id !== instanceId));
        addLog(`[Instance] Removed instance ${instanceId.substring(0, 8)}... from ${deviceId}`, LOG_TYPES.INFO);
      } else {
        addLog(`[Error] Failed to delete instance: ${response.statusText}`, LOG_TYPES.ERROR);
      }
    } catch (error) {
      console.error('Error deleting instance:', error);
      addLog(`[Error] Network error while deleting instance.`, LOG_TYPES.ERROR);
    } finally {
      setDeleteConfirmation(null);
    }
  };

  const openDeleteConfirmation = (instanceId, deviceId, productId) => {
    setDeleteConfirmation({ instanceId, deviceId, productId });
  };

  const cancelDelete = () => {
    setDeleteConfirmation(null);
  };

  const handleViewLogs = async (instanceId, productId) => {
    setLogsModal({ instanceId, productId, loading: true, logs: {} });
    try {
      const response = await fetch(`/api/deployments/${instanceId}/logs`);
      if (response.ok) {
        const data = await response.json();
        setLogsModal(prev => prev ? { ...prev, loading: false, logs: data.logs || {} } : null);
        addLog(`[Logs] Fetched logs for instance ${instanceId.substring(0, 8)}...`, LOG_TYPES.INFO);
      } else {
        setLogsModal(prev => prev ? { ...prev, loading: false, logs: { error: 'Failed to fetch logs' } } : null);
        addLog(`[Logs] Failed to fetch logs: ${response.statusText}`, LOG_TYPES.ERROR);
      }
    } catch (error) {
      console.error('Error fetching logs:', error);
      setLogsModal(prev => prev ? { ...prev, loading: false, logs: { error: error.message } } : null);
      addLog(`[Logs] Error fetching logs: ${error.message}`, LOG_TYPES.ERROR);
    }
  };

  const closeLogs = () => {
    setLogsModal(null);
  };

  const addLog = (msg, type = LOG_TYPES.INFO) => {
    setLogs(prev => [...prev, { 
      id: Date.now() + Math.random(), 
      time: new Date().toLocaleTimeString(), 
      msg, 
      type 
    }]);
  };

  const openCreateProduct = () => {
    setProductFormError('');
    setProductForm({
      id: '',
      name: '',
      description: '',
      deployJobs: { PROD: '', DEV: '' },
      deleteJob: ''
    });
    loadJobsCatalog();
    setIsCreateProductOpen(true);
  };

  const closeCreateProduct = () => {
    setIsCreateProductOpen(false);
  };

  const updateProductField = (field, value) => {
    setProductForm(prev => ({ ...prev, [field]: value }));
  };

  const updateDeployJobField = (env, value) => {
    setProductForm(prev => ({
      ...prev,
      deployJobs: { ...prev.deployJobs, [env]: value }
    }));
  };

  const handleCreateProduct = async () => {
    const trimmedId = productForm.id.trim();
    const trimmedName = productForm.name.trim();
    const trimmedDesc = productForm.description.trim();
    const deployJobs = Object.fromEntries(
      Object.entries(productForm.deployJobs)
        .map(([env, job]) => [env, job.trim()])
        .filter(([, job]) => job.length > 0)
    );
    const deleteJob = productForm.deleteJob.trim();

    if (!trimmedId || !trimmedName || Object.keys(deployJobs).length === 0 || !deleteJob) {
      setProductFormError('ID, name, deploy job, and delete job are required.');
      return;
    }

    try {
      const response = await fetch('/api/products', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: trimmedId,
          name: trimmedName,
          description: trimmedDesc,
          deploy_jobs: deployJobs,
          delete_job: deleteJob
        })
      });

      if (!response.ok) {
        const errData = await response.json().catch(() => ({}));
        const detail = errData.detail || response.statusText;
        setProductFormError(detail || 'Failed to create product.');
        addLog(`[Products] Create failed: ${detail}`, LOG_TYPES.ERROR);
        return;
      }

      const created = await response.json();
      const mappedProduct = {
        ...DEFAULT_PRODUCT_VALUES,
        id: created.id,
        name: created.name,
        description: created.description || 'No description available',
        deploy_jobs: created.deploy_jobs || deployJobs,
        delete_job: created.delete_job || deleteJob
      };

      setProducts(prev => [mappedProduct, ...prev]);
      addLog(`[Products] Created ${created.name} (${created.id}).`, LOG_TYPES.INFO);
      setIsCreateProductOpen(false);
    } catch (error) {
      console.error('Error creating product:', error);
      setProductFormError(error.message || 'Network error');
      addLog('[Products] Network error creating product.', LOG_TYPES.ERROR);
    }
  };

  const getLogColor = (type) => LOG_COLORS[type] || LOG_COLORS[LOG_TYPES.INFO];

  const groupedNodes = tailscaleNodes.reduce((acc, node) => {
    if (!acc[node.region]) acc[node.region] = [];
    acc[node.region].push(node);
    return acc;
  }, {});

  const groupedInstances = instances.reduce((acc, instance) => {
    if (!acc[instance.device_id]) acc[instance.device_id] = [];
    acc[instance.device_id].push(instance);
    return acc;
  }, {});

  const filteredProducts = products.filter(p => 
    p.name.toLowerCase().includes(filter.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-[#020617] text-slate-100 font-sans p-4 md:p-8 flex flex-col selection:bg-blue-500/30">
      <DashboardHeader 
        nodeCount={tailscaleNodes.length}

      />

      <main className="max-w-7xl w-full mx-auto grid grid-cols-1 lg:grid-cols-12 gap-6 flex-1">
        <div className="lg:col-span-8 space-y-6">
          {activeExecution && (
            <ActiveExecutionBanner execution={activeExecution} />
          )}

          <ProductCatalog
            products={filteredProducts}
            filter={filter}
            onFilterChange={setFilter}
            onDeploy={openDeployModal}
            onCreateProduct={openCreateProduct}
          />

          {Object.keys(groupedInstances).length > 0 && (
            <InstancesTreeView 
              groupedInstances={groupedInstances}
              onRefresh={loadInstances}
              onDelete={openDeleteConfirmation}
              onViewLogs={handleViewLogs}
            />
          )}

          <TailscaleTreeView groupedNodes={groupedNodes} />
        </div>

        <div className="lg:col-span-4 flex flex-col gap-6">
          <LiveConsole
            logs={logs}
            onClear={() => setLogs([])}
            logEndRef={logEndRef}
            getLogColor={getLogColor}
          />
        </div>
      </main>

      {isDeployModalOpen && (
        <DeployModal
          product={selectedProductForDeploy}
          nodes={tailscaleNodes}
          selectedHost={tempTargetHost}
          onSelectHost={setTempTargetHost}
          onConfirm={confirmDeployment}
          onClose={() => setIsDeployModalOpen(false)}
        />
      )}

      {deleteConfirmation && (
        <DeleteConfirmationModal
          instanceId={deleteConfirmation.instanceId}
          deviceId={deleteConfirmation.deviceId}
          productId={deleteConfirmation.productId}
          onConfirm={() => handleDeleteInstance(deleteConfirmation.instanceId, deleteConfirmation.deviceId, deleteConfirmation.productId)}
          onCancel={cancelDelete}
        />
      )}

      {logsModal && (
        <LogsModal
          instanceId={logsModal.instanceId}
          productId={logsModal.productId}
          loading={logsModal.loading}
          logs={logsModal.logs}
          onClose={closeLogs}
        />
      )}

      {isCreateProductOpen && (
        <CreateProductModal
          form={productForm}
          error={productFormError}
          jobsCatalog={jobsCatalog}
          jobsLoading={jobsLoading}
          jobsError={jobsError}
          onChange={updateProductField}
          onChangeDeployJob={updateDeployJobField}
          onClose={closeCreateProduct}
          onSubmit={handleCreateProduct}
        />
      )}

      <CustomScrollbarStyles />
    </div>
  );
}

const DashboardHeader = ({ nodeCount, onNewDevice }) => (
  <header className="max-w-7xl w-full mx-auto mb-8 flex flex-col md:flex-row md:items-center justify-between gap-4">
    <div className="flex items-center gap-4">
      <div className="p-3 bg-blue-600/10 rounded-2xl border border-blue-500/20 shadow-[0_0_20px_rgba(59,130,246,0.1)]">
        <Package className="text-blue-400 w-8 h-8" />
      </div>
      <div>
        <h1 className="text-2xl font-black tracking-tighter uppercase">
          ARK <span className="text-blue-500">Products</span>
        </h1>
        <div className="flex items-center gap-2 text-[10px] text-slate-500 font-bold uppercase tracking-widest">
          <span className="flex items-center gap-1 text-green-500">
            <ShieldCheck size={12}/> VPN: {nodeCount} Nodes
          </span>
          <span className="text-slate-800">|</span>
          <span className="flex items-center gap-1">
            <Network size={12}/> Mesh Topology Active
          </span>
        </div>
      </div>
    </div>


  </header>
);

const ActiveExecutionBanner = ({ execution }) => (
  <div className="bg-blue-600/5 border border-blue-500/30 rounded-2xl p-6 ring-1 ring-blue-500/10 relative overflow-hidden group">
    <div className="flex items-center justify-between relative z-10">
      <div className="flex items-center gap-5">
        <div className="w-12 h-12 bg-blue-600 rounded-2xl flex items-center justify-center shadow-[0_0_20px_rgba(37,99,235,0.4)]">
          <RefreshCw className="text-white animate-spin" size={24} />
        </div>
        <div>
          <h2 className="text-lg font-bold tracking-tight">Deploying to {execution.target}</h2>
          <p className="text-sm text-blue-400/70 font-medium tracking-wide italic">
            {execution.productName} • Build {execution.buildId}
          </p>
        </div>
      </div>
      <div className="bg-blue-500/20 px-4 py-1.5 rounded-full border border-blue-500/30">
        <span className="text-[10px] text-blue-300 font-black uppercase tracking-widest animate-pulse">
          {execution.status}...
        </span>
      </div>
    </div>
    <div className="absolute top-0 right-0 w-32 h-full bg-gradient-to-l from-blue-500/5 to-transparent group-hover:from-blue-500/10 transition-all duration-700" />
  </div>
);

const ProductCatalog = ({ products, filter, onFilterChange, onDeploy, onCreateProduct }) => (
  <section className="space-y-4">
    <div className="flex items-center justify-between px-2">
      <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-[0.3em] flex items-center gap-2">
        <Layers size={14} className="text-blue-500" /> Deployment Units
      </h3>
      <div className="flex items-center gap-3">
        <div className="relative">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600" />
          <input 
            type="text" 
            placeholder="Filter by name..."
            className="bg-slate-900/50 border border-slate-800 rounded-xl pl-9 pr-4 py-1.5 text-xs focus:ring-1 ring-blue-500 outline-none w-48 transition-all focus:w-64 placeholder:text-slate-700 text-white"
            value={filter}
            onChange={(e) => onFilterChange(e.target.value)}
          />
        </div>
        <button
          onClick={onCreateProduct}
          className="px-4 py-2 bg-slate-900/60 hover:bg-blue-600 text-white text-[10px] font-black uppercase rounded-xl transition-all"
        >
          New Product
        </button>
      </div>
    </div>

    <div className="grid gap-3">
      {products.map(product => (
        <ProductCard key={product.id} product={product} onDeploy={onDeploy} />
      ))}
    </div>
  </section>
);

const ProductCard = ({ product, onDeploy }) => (
  <div className="bg-slate-900/40 border border-slate-800/80 rounded-2xl p-6 hover:border-blue-500/30 transition-all group relative overflow-hidden">
    <div className="flex flex-col lg:flex-row justify-between gap-6 relative z-10">
      <div className="flex-1">
        <div className="flex items-center gap-3 mb-2">
          <h4 className="text-lg font-bold text-slate-100 group-hover:text-blue-400 transition-colors">
            {product.name}
          </h4>
          <span className={`text-[9px] px-2 py-0.5 rounded-md font-black tracking-widest ${
            product.env === 'PROD' 
              ? 'bg-red-500/10 text-red-500 border border-red-500/20' 
              : 'bg-yellow-500/10 text-yellow-500 border border-yellow-500/20'
          }`}>
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
          <span className="flex items-center gap-2 border-r border-slate-800 pr-4">
            <CheckCircle2 size={13} className={product.status === 'success' ? 'text-green-500' : 'text-slate-700'}/> 
            Last Artifact: {product.buildNum}
          </span>
          <span className="flex items-center gap-2">
            <Clock size={13}/> Updated: {product.lastRun}
          </span>
        </div>
      </div>

      <div className="flex items-center shrink-0">
        <button 
          onClick={() => onDeploy(product)}
          className="flex items-center justify-center gap-3 px-8 py-3 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-black uppercase transition-all shadow-[0_10px_30px_rgba(37,99,235,0.2)] active:scale-95 group-hover:translate-x-1"
        >
          <Play size={14} fill="white" /> New Deployment
        </button>
      </div>
    </div>
  </div>
);

const TailscaleTreeView = ({ groupedNodes }) => (
  <section className="bg-slate-900/20 border border-slate-800/60 rounded-3xl p-6">
    <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-[0.3em] mb-6 flex items-center gap-2">
      <GitBranch size={14} className="text-blue-500" /> Tailscale Network Tree
    </h3>
    
    <div className="font-mono text-[11px] space-y-4">
      <div className="flex items-center gap-2 text-blue-400 font-bold">
        <Globe size={14} />
        <span>tailnet.ark-servers.mesh</span>
      </div>

      {Object.keys(groupedNodes).map(region => (
        <RegionNode key={region} region={region} nodes={groupedNodes[region]} />
      ))}
    </div>
  </section>
);

const RegionNode = ({ region, nodes }) => (
  <div className="ml-4 border-l border-slate-800 pl-4 relative">
    <div className="absolute left-0 top-3 w-4 h-px bg-slate-800" />
    
    <div className="flex items-center gap-2 mb-3 text-slate-400 uppercase font-black tracking-widest text-[9px]">
      <span className="text-slate-700">├──</span>
      <Server size={12} className="text-slate-600" />
      <span>Region: {region}</span>
    </div>

    <div className="space-y-3 ml-4">
      {nodes.map((node, idx, arr) => (
        <DeviceNode 
          key={node.ip} 
          node={node} 
          isLast={idx === arr.length - 1}
        />
      ))}
    </div>
  </div>
);

const DeviceNode = ({ node, isLast }) => (
  <div className="flex items-start gap-3 relative group">
    <span className="text-slate-700 font-bold shrink-0">
      {isLast ? '└──' : '├──'}
    </span>
    
    <div className={`p-3 rounded-xl border border-slate-800/60 bg-slate-950/40 flex-1 flex items-center justify-between group-hover:border-slate-700 transition-colors ${
      node.status === 'offline' ? 'opacity-40 grayscale' : ''
    }`}>
      <div className="flex items-center gap-4">
        <div className={`w-2.5 h-2.5 rounded-full ${
          node.status === 'active' 
            ? 'bg-green-500 shadow-[0_0_10px_rgba(34,197,94,0.4)]' 
            : node.status === 'idle' 
              ? 'bg-yellow-500' 
              : 'bg-slate-800'
        }`} />
        <div>
          <p className="font-bold text-slate-200 tracking-tight group-hover:text-blue-400 transition-colors uppercase">
            {node.name}
          </p>
          <p className="text-[9px] text-slate-600 font-mono tracking-tighter">
            {node.ip} · {node.type}
          </p>
        </div>
      </div>
      
      <div className="flex items-center gap-4">
        <div className="text-right">
          <p className="text-[8px] text-slate-700 font-black uppercase mb-1">Latency</p>
          <div className="flex gap-0.5">
            {[1,2,3,4].map(b => (
              <div key={b} className={`w-1 h-2 rounded-full ${
                node.status === 'active' && b <= 3 ? 'bg-blue-500' : 'bg-slate-800'
              }`} />
            ))}
          </div>
        </div>
      </div>
    </div>
  </div>
);

const LiveConsole = ({ logs, onClear, logEndRef, getLogColor }) => (
  <section className="bg-black border border-slate-800/80 rounded-[2rem] overflow-hidden flex flex-col h-full min-h-[600px] shadow-2xl">
    <div className="bg-[#0f172a] px-6 py-4 border-b border-slate-800/60 flex items-center justify-between">
      <div className="flex items-center gap-3">
        <Terminal size={16} className="text-blue-500" />
        <span className="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">
          Deployment Stream
        </span>
      </div>
      <button 
        onClick={onClear}
        className="p-1.5 hover:bg-slate-800 rounded-lg text-slate-600 hover:text-red-400 transition-all"
      >
        <Eraser size={16} />
      </button>
    </div>
    
    <div className="p-6 font-mono text-[11px] overflow-y-auto flex-1 space-y-3 bg-[#020617] custom-scrollbar">
      {logs.map(log => (
        <div key={log.id} className="flex gap-4 leading-normal group">
          <span className="text-slate-800 shrink-0 select-none font-bold">[{log.time}]</span>
          <span className={`${getLogColor(log.type)} tracking-tight break-all`}>
            {log.type === LOG_TYPES.STAGE && '>> '} {log.msg}
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
);

const DeployModal = ({ product, nodes, selectedHost, onSelectHost, onConfirm, onClose }) => (
  <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-[#020617]/90 backdrop-blur-md">
    <div className="bg-slate-950 border border-slate-800 w-full max-w-lg rounded-[2rem] shadow-[0_0_100px_rgba(59,130,246,0.15)] overflow-hidden">
      <div className="px-8 py-6 border-b border-slate-800 flex items-center justify-between">
        <h2 className="font-black uppercase tracking-[0.2em] text-xs flex items-center gap-3">
          <Play size={18} className="text-blue-500" /> Confirm Deployment
        </h2>
        <button onClick={onClose} className="text-slate-500 hover:text-white transition-colors">
          <X size={24} />
        </button>
      </div>
      
      <div className="p-8 space-y-8">
        <DeployModalProduct product={product} />
        <DeployModalTargets 
          nodes={nodes}
          selectedHost={selectedHost}
          onSelectHost={onSelectHost}
        />
      </div>
      
      <div className="p-8 bg-black/40 border-t border-slate-800 flex gap-4">
        <button 
          onClick={onClose}
          className="flex-1 px-6 py-4 bg-slate-900 hover:bg-slate-800 rounded-2xl text-[11px] font-black uppercase transition-all tracking-widest text-slate-500"
        >
          Cancel
        </button>
        <button 
          onClick={onConfirm}
          className="flex-[2] px-6 py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl text-[11px] font-black uppercase transition-all shadow-[0_15px_40px_rgba(37,99,235,0.3)] flex items-center justify-center gap-3 tracking-[0.2em]"
        >
          Execute Pipeline <ArrowRight size={18} />
        </button>
      </div>
    </div>
  </div>
);

const DeployModalProduct = ({ product }) => (
  <div className="space-y-3">
    <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">
      Unit to Deploy
    </label>
    <div className="bg-slate-900/50 p-5 rounded-2xl border border-slate-800/80 flex items-center gap-4">
      <div className="p-3 bg-blue-500/10 rounded-xl">
        <Package className="text-blue-400" size={24}/>
      </div>
      <div>
        <p className="text-lg font-bold text-slate-100">{product.name}</p>
        <p className="text-xs text-blue-400/70 font-mono">{product.version} · {product.env}</p>
      </div>
    </div>
  </div>
);

const DeployModalTargets = ({ nodes, selectedHost, onSelectHost }) => (
  <div className="space-y-3">
    <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">
      Select Target Endpoint
    </label>
    <div className="grid gap-2 max-h-60 overflow-y-auto custom-scrollbar pr-2">
      {nodes.map(node => (
        <button
          key={node.ip}
          disabled={node.status === 'offline'}
          onClick={() => onSelectHost(node.ip)}
          className={`w-full p-4 rounded-2xl border text-left transition-all flex items-center justify-between ${
            selectedHost === node.ip 
              ? 'bg-blue-600 border-blue-500 shadow-lg shadow-blue-600/20' 
              : 'bg-slate-900/50 border-slate-800 hover:border-slate-700'
          } ${node.status === 'offline' ? 'opacity-20 cursor-not-allowed' : ''}`}
        >
          <div className="flex items-center gap-4">
            <div className={`w-2 h-2 rounded-full ${
              selectedHost === node.ip ? 'bg-white shadow-[0_0_10px_#fff]' : 'bg-green-500'
            }`} />
            <div>
              <p className={`text-xs font-black uppercase ${
                selectedHost === node.ip ? 'text-white' : 'text-slate-300'
              }`}>{node.name}</p>
              <p className={`text-[10px] font-mono ${
                selectedHost === node.ip ? 'text-blue-100' : 'text-slate-600'
              }`}>{node.ip}</p>
            </div>
          </div>
          {selectedHost === node.ip && <CheckCircle2 size={18} className="text-white" />}
        </button>
      ))}
    </div>
  </div>
);

const InstancesTreeView = ({ groupedInstances, onRefresh, onDelete, onViewLogs }) => (
  <div className="bg-slate-800/30 border border-slate-700 rounded-lg overflow-hidden">
    <div className="bg-slate-800/50 px-4 py-3 flex items-center justify-between border-b border-slate-700">
      <div className="flex items-center gap-2">
        <Server className="w-4 h-4 text-blue-400" />
        <span className="font-bold text-sm text-slate-300">Active Instances</span>
        <span className="text-xs text-slate-500 bg-slate-700 px-2 py-0.5 rounded-full">
          {Object.values(groupedInstances).reduce((sum, arr) => sum + arr.length, 0)}
        </span>
      </div>
      <button
        onClick={onRefresh}
        className="text-xs text-slate-400 hover:text-blue-400 transition-colors"
      >
        ↻ Refresh
      </button>
    </div>

    <div className="divide-y divide-slate-700">
      {Object.entries(groupedInstances).map(([deviceId, instances]) => (
        <div key={deviceId} className="p-4">
          <button className="flex items-center gap-2 w-full text-left hover:text-blue-400 transition-colors">
            <Terminal className="w-4 h-4 text-slate-500" />
            <span className="font-mono text-sm text-slate-300">{deviceId}</span>
            <span className="text-xs text-slate-500">({instances.length})</span>
          </button>

          <div className="ml-6 mt-3 space-y-2">
            {instances.map(instance => (
              <div
                key={instance.id}
                className="bg-slate-700/30 border border-slate-600 p-3 rounded-lg text-xs space-y-2"
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="font-mono text-blue-300">{instance.product_id}</span>
                  <div className="flex items-center gap-1">
                    <span className={`px-2 py-0.5 rounded text-xs font-bold ${
                      instance.status === 'running' ? 'bg-emerald-500/20 text-emerald-400' :
                      instance.status === 'provisioning' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-red-500/20 text-red-400'
                    }`}>
                      {instance.status}
                    </span>
                    <button
                      onClick={() => onViewLogs(instance.id, instance.product_id)}
                      className="p-1.5 hover:bg-blue-500/20 rounded transition-colors text-slate-400 hover:text-blue-400"
                      title="View logs"
                    >
                      <FileText className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={() => onDelete(instance.id, deviceId, instance.product_id)}
                      className="p-1.5 hover:bg-red-500/20 rounded transition-colors text-slate-400 hover:text-red-400"
                      title="Delete instance"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
                <div className="text-slate-400">
                  Environment: <span className="text-slate-300">{instance.environment}</span>
                </div>
                <div className="text-slate-400 break-all">
                  URL: <a href={instance.url} target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:text-blue-300">
                    {instance.url}
                  </a>
                </div>
                <div className="text-slate-500 text-[10px]">
                  ID: {instance.id.substring(0, 12)}
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  </div>
);

const DeleteConfirmationModal = ({ instanceId, deviceId, productId, onConfirm, onCancel }) => (
  <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div className="bg-slate-900 border border-slate-700 rounded-lg p-6 max-w-sm shadow-2xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 bg-red-500/10 rounded-lg">
          <AlertCircle className="w-5 h-5 text-red-400" />
        </div>
        <h2 className="text-lg font-bold text-red-400">Delete Instance</h2>
      </div>

      <p className="text-sm text-slate-300 mb-4">
        Are you sure you want to delete this instance?
        <br />
        <span className="text-slate-400 text-xs mt-2 block">
          Instance: <span className="font-mono text-slate-300">{instanceId.substring(0, 12)}...</span>
          <br />
          Device: <span className="font-mono text-slate-300">{deviceId}</span>
          <br />
          Product: <span className="font-mono text-slate-300">{productId}</span>
        </span>
      </p>

      <p className="text-xs text-slate-500 mb-6">
        This action <span className="text-red-400 font-bold">cannot be undone</span>.
      </p>

      <div className="flex gap-3">
        <button
          onClick={onCancel}
          className="flex-1 px-4 py-2 bg-slate-700/50 hover:bg-slate-700 border border-slate-600 rounded-lg text-sm font-medium transition-colors"
        >
          Cancel
        </button>
        <button
          onClick={onConfirm}
          className="flex-1 px-4 py-2 bg-red-600/20 hover:bg-red-600/30 border border-red-600/50 text-red-400 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2"
        >
          <Trash2 className="w-4 h-4" />
          Delete
        </button>
      </div>
    </div>
  </div>
);

const LogsModal = ({ instanceId, productId, loading, logs, onClose }) => (
  <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
    <div className="bg-slate-900 border border-slate-700 rounded-lg w-full max-w-2xl max-h-[80vh] shadow-2xl flex flex-col">
      <div className="bg-slate-800/50 px-6 py-4 border-b border-slate-700 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-2 bg-blue-500/10 rounded-lg">
            <FileText className="w-5 h-5 text-blue-400" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-slate-300">Build Logs</h2>
            <p className="text-xs text-slate-500">Product: <span className="font-mono">{productId}</span></p>
          </div>
        </div>
        <button
          onClick={onClose}
          className="p-2 hover:bg-slate-700 rounded transition-colors text-slate-400"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      <div className="flex-1 overflow-auto bg-slate-950 p-4 font-mono text-xs">
        {loading ? (
          <div className="flex items-center justify-center h-full">
            <div className="flex flex-col items-center gap-2">
              <Loader className="w-5 h-5 text-blue-400 animate-spin" />
              <span className="text-slate-400">Fetching logs...</span>
            </div>
          </div>
        ) : Object.keys(logs).length === 0 ? (
          <div className="text-slate-500 text-center py-8">No logs available</div>
        ) : (
          Object.entries(logs).map(([jobName, logContent]) => (
            <div key={jobName} className="mb-6">
              <div className="bg-slate-800/50 px-3 py-2 rounded border border-slate-700 mb-2">
                <span className="text-cyan-400 font-bold">Job: {jobName}</span>
              </div>
              <div className="bg-slate-950 border border-slate-800 rounded p-3 text-slate-300 whitespace-pre-wrap break-words max-h-64 overflow-y-auto">
                {typeof logContent === 'string' ? logContent : JSON.stringify(logContent, null, 2)}
              </div>
            </div>
          ))
        )}
      </div>

      <div className="bg-slate-800/50 px-6 py-3 border-t border-slate-700 flex justify-end">
        <button
          onClick={onClose}
          className="px-4 py-2 bg-slate-700/50 hover:bg-slate-700 border border-slate-600 rounded-lg text-sm font-medium transition-colors"
        >
          Close
        </button>
      </div>
    </div>
  </div>
);

const CreateProductModal = ({
  form,
  error,
  jobsCatalog,
  jobsLoading,
  jobsError,
  onChange,
  onChangeDeployJob,
  onClose,
  onSubmit
}) => {
  const [jobQuery, setJobQuery] = useState('');
  const normalizedJobs = (jobsCatalog || [])
    .map((job) => {
      if (typeof job === 'string') {
        return { key: job, label: job, value: job };
      }
      const label = job?.name || job?.id || '';
      return { key: job?.id || job?.name || label, label, value: label };
    })
    .filter(job => job.label);

  const filteredJobs = jobQuery.trim()
    ? normalizedJobs.filter(job => job.label.toLowerCase().includes(jobQuery.trim().toLowerCase()))
    : normalizedJobs;

  return (
  <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-[#020617]/90 backdrop-blur-md">
    <div className="bg-slate-950 border border-slate-800 w-full max-w-lg rounded-[2rem] shadow-[0_0_80px_rgba(59,130,246,0.15)] overflow-hidden">
      <div className="px-8 py-6 border-b border-slate-800 flex items-center justify-between">
        <h2 className="font-black uppercase tracking-[0.2em] text-xs flex items-center gap-3">
          <Package size={18} className="text-blue-500" /> New Product
        </h2>
        <button onClick={onClose} className="text-slate-500 hover:text-white transition-colors">
          <X size={24} />
        </button>
      </div>

      <div className="p-8 space-y-6">
        <div className="space-y-2">
          <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">
            Product ID
          </label>
          <input
            type="text"
            value={form.id}
            onChange={(e) => onChange('id', e.target.value)}
            placeholder="e.g. media-service"
            className="w-full bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-3 text-xs text-white placeholder:text-slate-700 focus:ring-1 ring-blue-500 outline-none"
          />
        </div>

        <div className="space-y-2">
          <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">
            Name
          </label>
          <input
            type="text"
            value={form.name}
            onChange={(e) => onChange('name', e.target.value)}
            placeholder="Product name"
            className="w-full bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-3 text-xs text-white placeholder:text-slate-700 focus:ring-1 ring-blue-500 outline-none"
          />
        </div>

        <div className="space-y-2">
          <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">
            Description
          </label>
          <textarea
            value={form.description}
            onChange={(e) => onChange('description', e.target.value)}
            placeholder="Short description"
            className="w-full bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-3 text-xs text-white placeholder:text-slate-700 focus:ring-1 ring-blue-500 outline-none min-h-[90px]"
          />
        </div>

        <div className="space-y-3">
          <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">
            Deploy Jobs by Environment
          </label>

          {!jobsError && normalizedJobs.length > 8 && (
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600" size={14} />
              <input
                type="text"
                value={jobQuery}
                onChange={(e) => setJobQuery(e.target.value)}
                placeholder="Filter jobs"
                className="w-full bg-slate-900/60 border border-slate-800 rounded-xl pl-9 pr-4 py-2 text-xs text-white placeholder:text-slate-700 focus:ring-1 ring-blue-500 outline-none"
              />
            </div>
          )}
          
          <div className="grid gap-3">
            <div className="flex items-center gap-3">
              <span className="text-[10px] font-black uppercase text-slate-500 w-12">PROD</span>
              {jobsError ? (
                <input
                  type="text"
                  value={form.deployJobs.PROD}
                  onChange={(e) => onChangeDeployJob('PROD', e.target.value)}
                  placeholder="jenkins-job-prod"
                  className="flex-1 bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-2 text-xs text-white placeholder:text-slate-700 focus:ring-1 ring-blue-500 outline-none"
                />
              ) : (
                <select
                  value={form.deployJobs.PROD}
                  onChange={(e) => onChangeDeployJob('PROD', e.target.value)}
                  className="flex-1 bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-2 text-xs text-white focus:ring-1 ring-blue-500 outline-none"
                  disabled={jobsLoading || normalizedJobs.length === 0}
                >
                  <option value="">Select job...</option>
                  {filteredJobs.map(job => (
                    <option key={job.key} value={job.value}>{job.label}</option>
                  ))}
                </select>
              )}
            </div>
            <div className="flex items-center gap-3">
              <span className="text-[10px] font-black uppercase text-slate-500 w-12">DEV</span>
              {jobsError ? (
                <input
                  type="text"
                  value={form.deployJobs.DEV}
                  onChange={(e) => onChangeDeployJob('DEV', e.target.value)}
                  placeholder="jenkins-job-dev"
                  className="flex-1 bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-2 text-xs text-white placeholder:text-slate-700 focus:ring-1 ring-blue-500 outline-none"
                />
              ) : (
                <select
                  value={form.deployJobs.DEV}
                  onChange={(e) => onChangeDeployJob('DEV', e.target.value)}
                  className="flex-1 bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-2 text-xs text-white focus:ring-1 ring-blue-500 outline-none"
                  disabled={jobsLoading || normalizedJobs.length === 0}
                >
                  <option value="">Select job...</option>
                  {filteredJobs.map(job => (
                    <option key={job.key} value={job.value}>{job.label}</option>
                  ))}
                </select>
              )}
            </div>
          </div>

          <div className="text-[10px] text-slate-500 flex items-center gap-1">
            {jobsLoading && 'Loading jobs from Jenkins...'}
            {!jobsLoading && jobsError && 'Jobs unavailable'}
            {!jobsLoading && !jobsError && normalizedJobs.length === 0 && 'No jobs found from Jenkins.'}
            {!jobsLoading && !jobsError && normalizedJobs.length > 0 && (
              <>
                <CheckCircle2 size={10} className="text-emerald-500" />
                <span>{normalizedJobs.length} job(s) available</span>
              </>
            )}
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-[10px] font-black text-slate-600 uppercase tracking-widest block">
            Delete Job (Admin Only)
          </label>
          {jobsError ? (
            <input
              type="text"
              value={form.deleteJob}
              onChange={(e) => onChange('deleteJob', e.target.value)}
              placeholder="jenkins-job-delete"
              className="w-full bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-3 text-xs text-white placeholder:text-slate-700 focus:ring-1 ring-blue-500 outline-none"
            />
          ) : (
            <select
              value={form.deleteJob}
              onChange={(e) => onChange('deleteJob', e.target.value)}
              className="w-full bg-slate-900/60 border border-slate-800 rounded-xl px-4 py-3 text-xs text-white focus:ring-1 ring-blue-500 outline-none"
              disabled={jobsLoading || normalizedJobs.length === 0}
            >
              <option value="">Select job...</option>
              {filteredJobs.map(job => (
                <option key={job.key} value={job.value}>{job.label}</option>
              ))}
            </select>
          )}
        </div>

        {error && (
          <div className="text-xs text-red-400 bg-red-500/10 border border-red-500/30 rounded-lg px-3 py-2">
            {error}
          </div>
        )}
      </div>

      <div className="p-8 bg-black/40 border-t border-slate-800 flex gap-4">
        <button 
          onClick={onClose}
          className="flex-1 px-6 py-4 bg-slate-900 hover:bg-slate-800 rounded-2xl text-[11px] font-black uppercase transition-all tracking-widest text-slate-500"
        >
          Cancel
        </button>
        <button 
          onClick={onSubmit}
          className="flex-[2] px-6 py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-2xl text-[11px] font-black uppercase transition-all shadow-[0_15px_40px_rgba(37,99,235,0.3)] flex items-center justify-center gap-3 tracking-[0.2em]"
        >
          Create Product
        </button>
      </div>
    </div>
  </div>
  );
};

const CustomScrollbarStyles = () => (
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
);
