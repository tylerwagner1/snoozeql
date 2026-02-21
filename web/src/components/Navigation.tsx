import { Link } from 'react-router-dom'
import { Activity, Database, Clock, Lightbulb, Cloud, FileText } from 'lucide-react'

const Navigation = () => (
  <nav className="flex items-center space-x-1 md:space-x-2 bg-slate-800/30 p-1.5 rounded-full">
    <Link to="/" className="flex items-center space-x-1.5 px-3 py-1.5 text-sm font-medium rounded-full transition-all duration-200 hover:bg-blue-500/20 hover:text-blue-400 text-slate-300">
      <Activity className="h-4 w-4" />
      <span className="hidden sm:inline">Dashboard</span>
    </Link>
    <Link to="/cloud-accounts" className="flex items-center space-x-1.5 px-3 py-1.5 text-sm font-medium rounded-full transition-all duration-200 hover:bg-purple-500/20 hover:text-purple-400 text-slate-300">
      <Cloud className="h-4 w-4" />
      <span className="hidden sm:inline">Accounts</span>
    </Link>
    <Link to="/instances" className="flex items-center space-x-1.5 px-3 py-1.5 text-sm font-medium rounded-full transition-all duration-200 hover:bg-blue-500/20 hover:text-blue-400 text-slate-300">
      <Database className="h-4 w-4" />
      <span className="hidden sm:inline">Instances</span>
    </Link>
    <Link to="/schedules" className="flex items-center space-x-1.5 px-3 py-1.5 text-sm font-medium rounded-full transition-all duration-200 hover:bg-blue-500/20 hover:text-blue-400 text-slate-300">
      <Clock className="h-4 w-4" />
      <span className="hidden sm:inline">Schedules</span>
    </Link>
    <Link to="/recommendations" className="flex items-center space-x-1.5 px-3 py-1.5 text-sm font-medium rounded-full transition-all duration-200 hover:bg-blue-500/20 hover:text-blue-400 text-slate-300">
      <Lightbulb className="h-4 w-4" />
      <span className="hidden sm:inline">Recommendations</span>
    </Link>
    <Link to="/audit-log" className="flex items-center space-x-1.5 px-3 py-1.5 text-sm font-medium rounded-full transition-all duration-200 hover:bg-blue-500/20 hover:text-blue-400 text-slate-300">
      <FileText className="h-4 w-4" />
      <span className="hidden sm:inline">Audit Log</span>
    </Link>
  </nav>
)

export default Navigation
