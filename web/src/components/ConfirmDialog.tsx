import { Dialog, DialogPanel, DialogTitle, DialogBackdrop } from '@headlessui/react'

interface ConfirmDialogProps {
  isOpen: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  message: string
  confirmText: string
  confirmVariant: 'danger' | 'warning' | 'success'
  loading?: boolean
}

const variantClasses = {
  danger: 'bg-red-600 hover:bg-red-500 shadow-red-500/20',
  warning: 'bg-yellow-600 hover:bg-yellow-500 shadow-yellow-500/20',
  success: 'bg-green-600 hover:bg-green-500 shadow-green-500/20',
}

export function ConfirmDialog({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText,
  confirmVariant,
  loading = false,
}: ConfirmDialogProps) {
  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <DialogBackdrop
        transition
        className="fixed inset-0 bg-black/50 backdrop-blur-sm duration-200 ease-out data-[closed]:opacity-0"
      />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel
          transition
          className="max-w-md w-full bg-slate-800 p-6 rounded-xl border border-slate-700 shadow-2xl duration-200 ease-out data-[closed]:scale-95 data-[closed]:opacity-0"
        >
          <DialogTitle className="text-lg font-semibold text-white">
            {title}
          </DialogTitle>
          <p className="mt-3 text-slate-300">{message}</p>
          <div className="mt-6 flex justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={onConfirm}
              disabled={loading}
              className={`px-4 py-2 text-white rounded-lg transition-all shadow-lg ${variantClasses[confirmVariant]} disabled:opacity-50`}
            >
              {loading ? 'Processing...' : confirmText}
            </button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  )
}

export default ConfirmDialog
