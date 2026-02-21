import React, { useState, useEffect } from 'react';
import { 
  Download, Shield, Cpu, Globe, CheckCircle, 
  Terminal, ArrowRight, ExternalLink, Package
} from 'lucide-react';

// Constants
const INSTALL_STEPS = {
  INITIALIZE: 1,
  AUTH_KEY: 2,
  GENERATE_SCRIPT: 3,
  READY: 4
};

const STEP_MESSAGES = {
  [INSTALL_STEPS.INITIALIZE]: 'Initializing instance environment...',
  [INSTALL_STEPS.AUTH_KEY]: 'Generating Tailscale AuthKey (Ephemeral)...',
  [INSTALL_STEPS.GENERATE_SCRIPT]: 'Generating deployment script...',
  [INSTALL_STEPS.READY]: 'Preparing download...'
};

const PRODUCT_ICONS = {
  media: Globe,
  db: Cpu,
  proxy: Shield,
  default: Package
};

const AUTH_KEY_CONFIG = {
  reusable: false,
  ephemeral: true,
  preauthorized: true,
  expiry_seconds: 3600
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

  const handleInstall = async (product) => {
    setSelectedProduct(product);
    setIsInstalling(true);
    setInstallStep(0);

    try {
      setInstallStep(INSTALL_STEPS.INITIALIZE);
      await sleep(1000);

      setInstallStep(INSTALL_STEPS.AUTH_KEY);
      const authKey = await generateAuthKey(product);
      await sleep(1000);

      setInstallStep(INSTALL_STEPS.GENERATE_SCRIPT);
      const installScript = generateInstallScript(product, authKey.auth_key);
      await sleep(1000);

      setInstallStep(INSTALL_STEPS.READY);
      setDeploymentData({
        product,
        authKey: authKey.auth_key,
        script: installScript,
        expiresAt: authKey.expires
      });
    } catch (error) {
      console.error('Deployment error:', error);
      alert('Error al preparar el despliegue: ' + error.message);
      setIsInstalling(false);
    }
  };

  const generateAuthKey = async (product) => {
    const response = await fetch('/api/tailscale/auth-keys', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        description: `Client deployment ${product.name}`,
        ...AUTH_KEY_CONFIG
      })
    });

    if (!response.ok) throw new Error('Failed to generate auth key');
    return response.json();
  };

  const generateInstallScript = (product, authKey) => `#!/bin/bash
# ARK Deployment Script - ${product.name}
# Generated: ${new Date().toISOString()}

echo "=== ARK Instance Deployment ==="
echo "Product: ${product.name}"
echo "Version: Latest"
echo ""

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo "Docker not installed. Install it first."; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "curl not installed. Install it first."; exit 1; }

# Install Tailscale if not present
if ! command -v tailscale >/dev/null 2>&1; then
    echo "Installing Tailscale..."
    curl -fsSL https://tailscale.com/install.sh | sh
fi

# Connect to Tailscale network
echo "Connecting to ARK mesh network..."
sudo tailscale up --authkey=${authKey}

# Deploy the product
echo "Deploying ${product.name}..."
# TODO: Add product-specific deployment commands here

echo ""
echo "Deployment complete!"
echo "Your instance is now connected to the ARK network."
`;

  const downloadScript = () => {
    const blob = new Blob([deploymentData.script], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `ark-deploy-${selectedProduct.id}.sh`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(deploymentData.script);
    alert('Script copiado al portapapeles');
  };

  const getProductIcon = (productId) => {
    const iconKey = Object.keys(PRODUCT_ICONS).find(key => productId.includes(key));
    const IconComponent = PRODUCT_ICONS[iconKey] || PRODUCT_ICONS.default;
    const colors = { media: 'text-blue-400', db: 'text-purple-400', proxy: 'text-emerald-400', default: 'text-slate-400' };
    const color = colors[iconKey] || colors.default;
    return <IconComponent className={`w-6 h-6 ${color}`} />;
  };

  const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

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
          onDownload={downloadScript}
          onCopy={copyToClipboard}
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
    <div className="hidden md:flex gap-8 text-sm font-medium text-slate-400">
      <a href="/" className="hover:text-blue-400 transition-colors">Products</a>
      <a href="/admin" className="hover:text-blue-400 transition-colors">Admin Panel</a>
    </div>
    <a href="/admin" className="px-4 py-2 bg-slate-900 border border-slate-700 rounded-lg text-sm hover:border-blue-500 transition-all">
      Dashboard
    </a>
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
      <a href="/admin" className="w-full sm:w-auto px-8 py-4 bg-slate-900 border border-slate-800 hover:bg-slate-800 text-slate-300 font-medium rounded-xl transition-all flex items-center justify-center gap-2">
        Admin Dashboard <ExternalLink className="w-4 h-4" />
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
    <p>No products available yet. Add them from the admin panel.</p>
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
      {Object.keys(product.jobs).map(env => (
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

const DeploymentModal = ({ isOpen, onClose, product, step, deploymentData, onDownload, onCopy }) => {
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
          <button onClick={onClose} className="text-slate-500 hover:text-white">✕</button>
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
              product={product}
              onDownload={onDownload}
              onCopy={onCopy}
            />
          )}
        </div>
      </div>
    </div>
  );
};

const DeploymentReady = ({ product, onDownload, onCopy }) => (
  <div className="mt-8 pt-6 border-t border-slate-800 animate-in fade-in slide-in-from-bottom-2 duration-700">
    <div className="bg-emerald-500/10 border border-emerald-500/20 p-4 rounded-xl mb-4">
      <div className="flex items-center justify-between mb-2">
        <p className="text-emerald-400 font-bold text-sm">Deployment Script Ready!</p>
        <span className="text-xs text-slate-500">Expires in 1h</span>
      </div>
      <p className="text-slate-400 text-xs mb-3">
        Download and execute this script on your target machine to deploy the instance.
      </p>
      <div className="flex gap-2">
        <button 
          onClick={onDownload}
          className="flex-1 px-4 py-2 bg-emerald-600 text-white rounded-lg text-xs font-bold hover:bg-emerald-500 transition-colors flex items-center justify-center gap-2"
        >
          <Download className="w-4 h-4" /> Download Script
        </button>
        <button 
          onClick={onCopy}
          className="px-4 py-2 bg-slate-700 text-white rounded-lg text-xs font-bold hover:bg-slate-600 transition-colors"
        >
          Copy
        </button>
      </div>
    </div>
    <div className="text-xs text-slate-500 space-y-1">
      <p>Instructions:</p>
      <ol className="list-decimal list-inside space-y-1 ml-2">
        <li>Download the script to your target machine</li>
        <li>Make it executable: <code className="bg-slate-800 px-1 rounded">chmod +x ark-deploy-{product.id}.sh</code></li>
        <li>Run it: <code className="bg-slate-800 px-1 rounded">sudo ./ark-deploy-{product.id}.sh</code></li>
      </ol>
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
        <a href="/admin" className="hover:text-blue-400">Admin</a>
      </div>
    </div>
  </footer>
);
