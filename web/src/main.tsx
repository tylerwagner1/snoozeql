import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import App from './App'
import Dashboard from './pages/Dashboard'
import CloudAccountsPage from './pages/CloudAccountsPage'
import InstancesPage from './pages/InstancesPage'
import InstanceDetailPage from './pages/InstanceDetailPage'
import SchedulesPage from './pages/SchedulesPage'
import ScheduleNewPage from './pages/ScheduleNewPage'
import ScheduleEditPage from './pages/ScheduleEditPage'
import RecommendationsPage from './pages/RecommendationsPage'
import AuditLogPage from './pages/AuditLogPage'
import './index.css'

const root = createRoot(document.getElementById('root')!)
root.render(
  <BrowserRouter>
    <Toaster 
      position="top-right"
      toastOptions={{
        style: {
          background: '#1e293b',
          color: '#fff',
          border: '1px solid #334155',
        },
        error: {
          iconTheme: {
            primary: '#ef4444',
            secondary: '#fff',
          },
        },
      }}
    />
    <Routes>
      <Route path="/" element={<App />}>
        <Route index element={<Dashboard />} />
        <Route path="cloud-accounts" element={<CloudAccountsPage />} />
        <Route path="instances" element={<InstancesPage />} />
        <Route path="instances/:id" element={<InstanceDetailPage />} />
        <Route path="schedules" element={<SchedulesPage />} />
        <Route path="schedules/new" element={<ScheduleNewPage />} />
        <Route path="schedules/:id" element={<ScheduleEditPage />} />
        <Route path="recommendations" element={<RecommendationsPage />} />
        <Route path="audit-log" element={<AuditLogPage />} />
      </Route>
    </Routes>
  </BrowserRouter>,
)
