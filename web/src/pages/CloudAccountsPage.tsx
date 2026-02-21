import { useState, useEffect } from 'react'
import { Plus, Trash2, Cloud, CloudOff, Check, X, RefreshCw, AlertCircle } from 'lucide-react'
import api, { CloudAccount } from '../lib/api'

const CloudAccountsPage = () => {
  const [accounts, setAccounts] = useState<CloudAccount[]>([])
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [form, setForm] = useState({
    name: '',
    provider: 'aws' as 'aws' | 'gcp',
    accessKeyId: '',
    secretAccessKey: '',
    gcpProjectId: '',
    gcpServiceKey: '',
    regions: 'us-east-1,us-west-2',
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  
  const connectionStatusColors: Record<string, string> = {
    connected: 'bg-green-500/10 text-green-400 border-green-500/30',
    syncing: 'bg-blue-500/10 text-blue-400 border-blue-500/30 animate-pulse',
    failed: 'bg-red-500/10 text-red-400 border-red-500/30',
    unknown: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
  }

  const loadAccounts = async () => {
    setLoading(true)
    try {
      const accounts = await api.getCloudAccounts()
      setAccounts(accounts)
    } catch (err) {
      console.error('Failed to load accounts:', err)
      setError('Failed to load cloud accounts')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadAccounts()
  }, [])
  
  const SkeletonCard = () => (
    <div className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700 animate-pulse">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-slate-700 rounded-lg" />
          <div>
            <div className="h-4 w-24 bg-slate-700 rounded mb-2" />
            <div className="h-3 w-12 bg-slate-700 rounded" />
          </div>
        </div>
        <div className="w-8 h-8 bg-slate-700 rounded" />
      </div>
      <div className="space-y-2">
        <div className="h-3 w-20 bg-slate-700 rounded" />
        <div className="h-3 w-32 bg-slate-700 rounded" />
      </div>
    </div>
  )

      handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setLoading(true)
        setError('')

        try {
          const credentials: { [key: string]: string } = {}
          
          if (form.provider === 'aws') {
            credentials.aws_access_key_id = form.accessKeyId
            credentials.aws_secret_access_key = form.secretAccessKey
            credentials.region = form.regions.split(',')[0].trim()
          } else {
            credentials.gcp_project_id = form.gcpProjectId
            credentials.gcp_service_account_key = form.gcpServiceKey
          }

          await api.createCloudAccount({
            name: form.name,
            provider: form.provider,
            regions: form.regions.split(',').map(r => r.trim()),
            credentials,
          })

          setIsModalOpen(false)
          setForm({ name: '', provider: 'aws', accessKeyId: '', secretAccessKey: '', gcpProjectId: '', gcpServiceKey: '', regions: 'us-east-1,us-west-2' })
          setError('')
          loadAccounts()
        } catch (err) {
          setError('Failed to create cloud account')
        } finally {
          setLoading(false)
        }
      }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to disconnect this account?')) return

    try {
      await api.deleteCloudAccount(id)
      loadAccounts()
    } catch (err) {
      setError('Failed to delete cloud account')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Cloud Accounts</h1>
          <p className="text-sm text-slate-400 mt-1">
            Connect your AWS or GCP accounts to discover and manage database instances
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 text-white text-sm font-medium rounded-lg shadow-lg shadow-blue-500/20 transition-all"
        >
          <Plus className="h-4 w-4" />
          Connect Account
        </button>
      </div>

      {accounts.length === 0 ? (
        <div className="text-center py-20 bg-slate-800/30 rounded-2xl border border-slate-700/50">
          <CloudOff className="h-16 w-16 text-slate-500 mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-white mb-2">No connected accounts</h3>
          <p className="text-slate-400 max-w-md mx-auto mb-6">
            Connect your cloud provider account to automatically discover database instances, track costs, and apply sleep schedules
          </p>
          <button
            onClick={() => setIsModalOpen(true)}
            className="px-6 py-2 bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 text-white font-medium rounded-lg shadow-lg shadow-blue-500/20 transition-all"
          >
            Connect Your First Account
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              <SkeletonCard />
              <SkeletonCard />
              <SkeletonCard />
            </div>
          ) : (
            accounts.map((account) => (
              <div key={account.id} className="bg-slate-800/50 rounded-xl p-6 shadow-lg border border-slate-700 hover:border-slate-600 transition-all">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    {account.provider === 'aws' ? (
                      <div className="p-2 bg-gradient-to-br from-yellow-500 to-orange-600 rounded-lg shadow-lg shadow-yellow-500/20">
                        <Cloud className="h-6 w-6 text-white" />
                      </div>
                    ) : (
                      <div className="p-2 bg-gradient-to-br from-red-500 to-pink-600 rounded-lg shadow-lg shadow-red-500/20">
                        <Cloud className="h-6 w-6 text-white" />
                      </div>
                    )}
                    <div>
                      <h3 className="font-semibold text-white">{account.name}</h3>
                      <span className="text-xs text-slate-400 uppercase font-medium">{account.provider}</span>
                    </div>
                  </div>
                  <button
                    onClick={() => handleDelete(account.id)}
                    className="p-2 text-slate-400 hover:text-red-500 transition-colors rounded-lg hover:bg-red-500/10"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-sm text-slate-300">
                    <span className={`px-2 py-0.5 text-xs rounded-full border ${
                      connectionStatusColors[account.connection_status || 'unknown']
                    }`}>
                      {account.connection_status === 'connected' && <Check className="inline h-3 w-3 mr-1" />}
                      {account.connection_status === 'syncing' && <RefreshCw className="inline h-3 w-3 mr-1 animate-spin" />}
                      {account.connection_status === 'failed' && <AlertCircle className="inline h-3 w-3 mr-1" />}
                      {(account.connection_status || 'unknown').charAt(0).toUpperCase() + (account.connection_status || 'unknown').slice(1)}
                    </span>
                  </div>
                  {account.connection_status === 'failed' && account.last_error && (
                    <p className="text-xs text-red-400 mt-1 truncate" title={account.last_error}>
                      {account.last_error.substring(0, 50)}...
                    </p>
                  )}
                  {account.regions && (
                    <div className="flex items-center gap-2 text-sm text-slate-400">
                      <span className="text-slate-500">Regions:</span>
                      <span className="bg-slate-900/50 px-2 py-1 rounded text-xs">{account.regions.join(', ')}</span>
                    </div>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 z-50">
          <div className="bg-slate-900 rounded-2xl border border-slate-700 shadow-2xl max-w-md w-full p-6 animate-in fade-in zoom-in duration-200">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-white">Connect Cloud Account</h2>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-slate-400 hover:text-white transition-colors"
              >
                <X className="h-6 w-6" />
              </button>
            </div>

            {loading ? null : error && (
              <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm flex items-center gap-2">
                <X className="h-4 w-4" />
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">Account Name</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="e.g., Production AWS Account"
                  required
                  className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
                />
                <p className="text-xs text-slate-500 mt-1">
                  A friendly name to identify this account in SnoozeQL
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">Provider</label>
                <div className="grid grid-cols-2 gap-3">
                  <button
                    type="button"
                    onClick={() => setForm({ ...form, provider: 'aws' })}
                    className={`flex items-center justify-center gap-2 px-4 py-3 rounded-lg border-2 transition-all ${
                      form.provider === 'aws'
                        ? 'border-yellow-500 bg-yellow-500/10 text-yellow-500'
                        : 'border-slate-700 bg-slate-800 text-slate-400 hover:border-slate-600'
                    }`}
                  >
                    <Cloud className="h-5 w-5" />
                    <span className="font-medium">AWS</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => setForm({ ...form, provider: 'gcp' })}
                    className={`flex items-center justify-center gap-2 px-4 py-3 rounded-lg border-2 transition-all ${
                      form.provider === 'gcp'
                        ? 'border-red-500 bg-red-500/10 text-red-500'
                        : 'border-slate-700 bg-slate-800 text-slate-400 hover:border-slate-600'
                    }`}
                  >
                    <Cloud className="h-5 w-5" />
                    <span className="font-medium">GCP</span>
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  {form.provider === 'aws' ? 'AWS Access Key ID' : 'GCP Project ID'}
                </label>
                <input
                  type="text"
                  value={form.provider === 'aws' ? form.accessKeyId : form.gcpProjectId}
                  onChange={(e) => setForm({ ...form, [form.provider === 'aws' ? 'accessKeyId' : 'gcpProjectId']: e.target.value })}
                  placeholder={form.provider === 'aws' ? 'e.g., AKIAIOSFODNN7EXAMPLE' : 'e.g., my-gcp-project-123'}
                  required
                  className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono text-sm"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  {form.provider === 'aws' ? 'AWS Secret Access Key' : 'Service Account JSON Key'}
                </label>
                <input
                  type="password"
                  value={form.provider === 'aws' ? form.secretAccessKey : form.gcpServiceKey}
                  onChange={(e) => setForm({ ...form, [form.provider === 'aws' ? 'secretAccessKey' : 'gcpServiceKey']: e.target.value })}
                  placeholder={form.provider === 'aws' ? '40-character secret (starts with special characters)' : 'Paste GCP service account JSON key'}
                  required
                  className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono text-sm"
                />
                <p className="text-xs text-slate-500 mt-1">
                  {form.provider === 'aws'
                    ? 'AWS Secret Access Keys are 40 characters and don\'t start with "AKIA"'
                    : 'You can paste the entire JSON key file content here'}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">AWS Regions (comma-separated)</label>
                <input
                  type="text"
                  value={form.regions}
                  onChange={(e) => setForm({ ...form, regions: e.target.value })}
                  placeholder="us-east-1,us-west-2"
                  className="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
                />
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600 text-white font-medium rounded-lg shadow-lg shadow-blue-500/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                {loading ? (
                  <>
                    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    Connecting...
                  </>
                ) : (
                  <>
                    <Cloud className="h-4 w-4" />
                    Connect Account
                  </>
                )}
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}

export default CloudAccountsPage
