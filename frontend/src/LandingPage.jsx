import React, { useState, useEffect } from 'react';
import { 
  Download, Shield, Cpu, Globe, CheckCircle, 
  Terminal, ArrowRight, Package, ExternalLink, X
} from 'lucide-react';

// Constants
const INSTALL_STEPS = {
  INITIALIZE: 1,
  CREATING: 2,
  READY: 3
};

const STEP_MESSAGES = {
  [INSTALL_STEPS.INITIALIZE]: 'Initializing instance environment...',
  [INSTALL_STEPS.CREATING]: 'Creating Docker containers...',
  [INSTALL_STEPS.READY]: 'Instance ready!'
};

const PRODUCT_ICONS = {
  media: Globe,
  db: Cpu,
  proxy: Shield,
  default: Package
};

export default function LandingPage() {
  const [products, setProducts] = useState([]);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [isInstalling, setIsInstalling] = useState(false);
  const [installStep, setInstallStep] = useState(0);
  const [deploymentData, setDeploymentData] = useState(null);

  useEffect(() => {
    loadProducts();
  }, []);

  useEffect(() => {
    if (isInstalling && installStep > 0 && installStep < INSTALL_STEPS.READY) {
      const timer = setTimeout(() => setInstallStep(prev => prev + 1), 1500);
      return () => clearTimeout(timer);
    }
  }, [isInstalling, installStep]);

  const loadProducts = async () => {
    try {
      const response = await fetch('/api/products');
      const data = await response.json();
      setProducts(data.products || []);
    } catch (error) {
      console.error('Error loading products:', error);
    }
  };

  const pickTargetHost = (device) => {
    const addrs = Array.isArray(device?.addresses) ? device.addresses : [];
    const ts = addrs.find(a => typeof a === "string" && a.startsWith("100."));
    return ts || "";
  };

  const isDeviceAvailable = (d) => {
    if (!d) return false;
    if (d.online === true) return true;
    if (d.active === true) return true;
    if (d.status === "active") return true;
    if (d.state === "active") return true;
    return Array.isArray(d.addresses) && d.addresses.some(a => typeof a === "string" && a.startsWith("100."));
  };

  const handleInstall = async (product) => {
    setSelectedProduct(product);
    setIsInstalling(true);
    setInstallStep(0);

    try {
      setInstallStep(INSTALL_STEPS.INITIALIZE);

      // Obtener un nodo disponible de Tailscale
      const devicesResponse = await fetch('/api/tailscale/devices');
      if (!devicesResponse.ok) {
        throw new Error(`Devices fetch failed: ${devicesResponse.status}`);
      }

      const devicesData = await devicesResponse.json();
      const availableDevice = (devicesData.devices || []).find(isDeviceAvailable);
      
      if (!availableDevice) {
        throw new Error('No available devices in network');
      }

      const targetHost = pickTargetHost(availableDevice);
      if (!targetHost) {
        throw new Error('No Tailscale IPv4 (100.x) found for selected device');
      }

      setInstallStep(INSTALL_STEPS.CREATING);

      // POST real a /api/deployments
      const deployResponse = await fetch('/api/deployments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          product_id: product.id,
          environment: 'PROD',
          target_host: targetHost
        })
      });

      if (!deployResponse.ok) {
        throw new Error(`Deployment failed: ${deployResponse.statusText}`);
      }

      const deployData = await deployResponse.json();

      setInstallStep(INSTALL_STEPS.READY);
      setDeploymentData({
        instanceId: deployData.instance_id,
        accessUrl: deployData.url,
        product
      });
    } catch (error) {
      console.error('Installation error:', error);
      alert('Failed to create instance: ' + error.message);
      setIsInstalling(false);
    }
  };

  const getProductIcon = (productId) => {
    const iconKey = Object.keys(PRODUCT_ICONS).find(key => productId.includes(key));
    const IconComponent = PRODUCT_ICONS[iconKey] || PRODUCT_ICONS.default;
    const colors = { media: 'text-blue-400', db: 'text-purple-400', proxy: 'text-emerald-400', default: 'text-slate-400' };
    const color = colors[iconKey] || colors.default;
    return <IconComponent className={`w-6 h-6 ${color}`} />;
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-200 font-sans selection:bg-blue-500/30">
      <BackgroundDecor />
      <Navigation />
      <HeroSection />
      <ProductGrid 
        products={products} 
        onInstall={handleInstall}
        getProductIcon={getProductIcon}
      />
      
      {isInstalling && (
        <DeploymentModal
          isOpen={isInstalling}
          onClose={() => setIsInstalling(false)}
          product={selectedProduct}
          step={installStep}
          deploymentData={deploymentData}
        />
      )}
      
      <Footer />
    </div>
  );
}

const BackgroundDecor = () => (
  <div className="fixed inset-0 overflow-hidden pointer-events-none">
    <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-900/10 rounded-full blur-[120px]" />
    <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-900/10 rounded-full blur-[120px]" />
  </div>
);

const Navigation = () => (
  <nav className="relative z-10 flex items-center justify-between px-8 py-6 max-w-7xl mx-auto border-b border-slate-800/50">
    <div className="flex items-center gap-2">
      <div className="w-8 h-8 bg-blue-600 rounded flex items-center justify-center shadow-lg shadow-blue-500/20">
        <span className="font-bold text-white tracking-tighter italic">A</span>
      </div>
      <span className="text-xl font-bold tracking-tight text-white uppercase">
        Ark <span className="text-blue-500 font-medium lowercase italic">store</span>
      </span>
    </div>
  </nav>
);

const HeroSection = () => (
  <header className="relative z-10 max-w-7xl mx-auto px-8 py-24 text-center">
    <h1 className="text-5xl md:text-7xl font-extrabold text-white mb-6 tracking-tight">
      Deploy Your Infrastructure <br />
      <span className="bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent">
        In One Click.
      </span>
    </h1>
    <p className="max-w-2xl mx-auto text-lg text-slate-400 mb-10 leading-relaxed">
      ARK provides isolated, secure instances connected via Tailscale. 
      Manage everything from your private dashboard while we handle the mesh.
    </p>
    <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
      <a href="#products" className="w-full sm:w-auto px-8 py-4 bg-blue-600 hover:bg-blue-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-blue-600/20 flex items-center justify-center gap-2">
        Explore Products <ArrowRight className="w-5 h-5" />
      </a>
    </div>
  </header>
);

const ProductGrid = ({ products, onInstall, getProductIcon }) => (
  <main id="products" className="relative z-10 max-w-7xl mx-auto px-8 pb-32">
    <div className="flex items-center justify-between mb-12">
      <h2 className="text-2xl font-bold text-white">Available Modules</h2>
      <div className="h-px flex-1 bg-slate-800 mx-8 opacity-50" />
    </div>
    
    {products.length === 0 ? (
      <EmptyState />
    ) : (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {products.map(product => (
          <ProductCard 
            key={product.id}
            product={product}
            onInstall={onInstall}
            icon={getProductIcon(product.id)}
          />
        ))}
      </div>
    )}
  </main>
);

const EmptyState = () => (
  <div className="text-center py-16 text-slate-500">
    <Package className="w-12 h-12 mx-auto mb-4 opacity-50" />
    <p>No products available yet. Check back soon!</p>
  </div>
);

const ProductCard = ({ product, onInstall, icon }) => (
  <div className="group relative bg-slate-900/50 border border-slate-800 p-8 rounded-2xl hover:border-blue-500/50 transition-all hover:translate-y-[-4px]">
    <div className="mb-6 bg-slate-800 w-12 h-12 rounded-xl flex items-center justify-center group-hover:bg-blue-500/10 transition-colors">
      {icon}
    </div>
    <div className="flex items-center justify-between mb-2">
      <h3 className="text-xl font-bold text-white">{product.name}</h3>
      <span className="text-xs font-mono text-slate-500 bg-slate-800 px-2 py-1 rounded">latest</span>
    </div>
    <p className="text-slate-400 text-sm leading-relaxed mb-6">{product.description}</p>
    <div className="flex flex-wrap gap-2 mb-8">
      {Object.keys(product.deploy_jobs || product.jobs || {}).map(env => (
        <span key={env} className="text-[10px] uppercase tracking-wider font-bold text-slate-500 border border-slate-800 px-2 py-0.5 rounded">
          {env}
        </span>
      ))}
    </div>
    <button 
      onClick={() => onInstall(product)}
      className="w-full py-3 bg-slate-800 hover:bg-blue-600 text-white font-semibold rounded-xl transition-all flex items-center justify-center gap-2"
    >
      <Download className="w-4 h-4" /> Deploy Instance
    </button>
  </div>
);

const DeploymentModal = ({ isOpen, onClose, product, step, deploymentData }) => {
  if (!isOpen) return null;

  const getStepIcon = (currentStep, targetStep) => {
    if (currentStep > targetStep) return <CheckCircle className="w-4 h-4 text-emerald-500" />;
    if (currentStep === targetStep) return <div className="w-4 h-4 rounded-full border border-slate-600 border-t-blue-400 animate-spin" />;
    return <div className="w-4 h-4 rounded-full border border-slate-700" />;
  };

  const getStepClass = (currentStep, targetStep) => {
    if (currentStep >= targetStep) return 'opacity-100';
    return 'opacity-20';
  };

  const getTextClass = (currentStep, targetStep) => {
    if (currentStep > targetStep) return 'text-emerald-400';
    return 'text-slate-300';
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm">
      <div className="bg-slate-900 border border-slate-800 w-full max-w-xl rounded-2xl overflow-hidden shadow-2xl">
        <div className="bg-slate-800 px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Terminal className="w-4 h-4 text-blue-400" />
            <span className="text-sm font-mono text-slate-300">ark_deployment_orchestrator</span>
          </div>
          <button onClick={onClose} className="text-slate-500 hover:text-white transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>
        
        <div className="p-8">
          <div className="flex items-center gap-4 mb-8">
            <div className="w-12 h-12 bg-blue-500/10 rounded-full flex items-center justify-center">
              {step < INSTALL_STEPS.READY ? (
                <Download className="w-6 h-6 text-blue-400 animate-bounce" />
              ) : (
                <CheckCircle className="w-6 h-6 text-emerald-400" />
              )}
            </div>
            <div>
              <h4 className="font-bold text-white">Deploying {product?.name}</h4>
              <p className="text-xs text-slate-500 font-mono italic">
                ID: {Math.random().toString(16).substring(7)}
              </p>
            </div>
          </div>

          <div className="space-y-4 font-mono text-xs">
            {Object.entries(STEP_MESSAGES).map(([stepNum, message]) => (
              <div key={stepNum} className={`flex items-center gap-3 transition-opacity ${getStepClass(step, parseInt(stepNum))}`}>
                {getStepIcon(step, parseInt(stepNum))}
                <span className={getTextClass(step, parseInt(stepNum))}>{message}</span>
              </div>
            ))}
          </div>

          {step === INSTALL_STEPS.READY && deploymentData && (
            <DeploymentReady 
              deploymentData={deploymentData}
            />
          )}
        </div>
      </div>
    </div>
  );
};

const DeploymentReady = ({ deploymentData }) => (
  <div className="mt-8 pt-6 border-t border-slate-800 animate-in fade-in slide-in-from-bottom-2 duration-700">
    <div className="bg-emerald-500/10 border border-emerald-500/20 p-6 rounded-xl mb-6">
      <div className="flex items-center justify-between mb-3">
        <p className="text-emerald-400 font-bold text-base flex items-center gap-2">
          <CheckCircle className="w-5 h-5" />
          Instance Deployed!
        </p>
        <span className="text-xs text-slate-400 font-mono bg-slate-800 px-2 py-1 rounded">
          ID: {deploymentData.instanceId.substring(0, 8)}...
        </span>
      </div>
      <p className="text-slate-400 text-sm mb-4">
        Your {deploymentData.product.name} instance is running and ready to use.
      </p>
      
      <div className="bg-slate-800/50 border border-slate-700 p-3 rounded-lg mb-4">
        <p className="text-xs text-slate-500 mb-1">Access URL:</p>
        <p className="text-sm font-mono text-blue-300 break-all">{deploymentData.accessUrl}</p>
      </div>

      <a
        href={deploymentData.accessUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="block w-full px-4 py-3 bg-gradient-to-r from-emerald-600 to-emerald-500 text-white rounded-lg text-sm font-bold hover:from-emerald-500 hover:to-emerald-400 transition-all shadow-lg shadow-emerald-500/20 text-center flex items-center justify-center gap-2"
      >
        <ExternalLink className="w-4 h-4" />
        Open Instance in New Tab
      </a>
    </div>

    <div className="text-xs text-slate-500 space-y-2">
      <p className="font-bold text-slate-400">Instance Details:</p>
      <ul className="space-y-1 ml-2">
        <li>Product: <span className="text-slate-300">{deploymentData.product.name}</span></li>
        <li className="flex items-center gap-2">
          Status: <span className="text-emerald-400 flex items-center gap-1">
            <span className="w-2 h-2 bg-emerald-400 rounded-full"></span> Running
          </span>
        </li>
        <li>Instance ID: <code className="bg-slate-800 px-1 rounded text-slate-300">{deploymentData.instanceId}</code></li>
      </ul>
    </div>
  </div>
);

const Footer = () => (
  <footer className="relative z-10 border-t border-slate-900 bg-slate-950 py-12">
    <div className="max-w-7xl mx-auto px-8 flex flex-col md:flex-row justify-between items-center gap-6">
      <div className="text-slate-500 text-sm">
        © 2026 ARK Products & Deployments. All rights reserved.
      </div>
      <div className="flex gap-6 text-slate-500 text-sm">
        <a href="#" className="hover:text-blue-400">Terms</a>
        <a href="#" className="hover:text-blue-400">Privacy</a>
      </div>
    </div>
  </footer>
);
