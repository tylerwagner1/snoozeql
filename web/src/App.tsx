import { Outlet } from 'react-router-dom'
import Navigation from './components/Navigation'

export default function App() {
  return (
    <div className="min-h-screen bg-background text-foreground font-sans selection:bg-blue-500 selection:text-white">
      {/* Header */}
      <header className="border-b border-border bg-card/50 backdrop-blur-md sticky top-0 z-50">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center">
            <span className="text-xl font-bold bg-gradient-to-r from-primary to-purple-400 bg-clip-text text-transparent">
              SnoozeQL
            </span>
          </div>
          <Navigation />
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}
