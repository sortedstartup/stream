import React, { useState } from 'react'
import { X, AlertCircle } from 'react-feather'

export const CreateWorkspaceModal = ({ isOpen, onClose, onSubmit }) => {
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  if (!isOpen) return null

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setIsSubmitting(true)

    const formData = new FormData(e.target)
    const name = formData.get('name')
    const description = formData.get('description')
    
    if (name) {
      try {
        const result = await onSubmit(name, description || '')
        if (result.success) {
          // Success - modal will be closed by parent
          setError('')
        } else {
          setError(result.error || 'Failed to create workspace. Please try again.')
        }
      } catch (err) {
        setError('Failed to create workspace. Please try again.')
      }
    }
    
    setIsSubmitting(false)
  }

  const handleClose = () => {
    setError('')
    setIsSubmitting(false)
    onClose()
  }

  return (
    <div className="modal modal-open">
      <div className="modal-box relative max-w-md">
        <button 
          className="btn btn-sm btn-circle btn-ghost absolute right-2 top-2"
          onClick={handleClose}
        >
          <X className="w-4 h-4" />
        </button>

        <h3 className="font-bold text-lg mb-4">Create New Workspace</h3>
        
        {error && (
          <div className="alert alert-error mb-4">
            <AlertCircle className="stroke-current shrink-0 h-4 w-4" />
            <span className="text-sm">{error}</span>
            <button 
              className="btn btn-sm btn-ghost btn-circle ml-auto"
              onClick={() => setError('')}
            >
              <X className="w-3 h-3" />
            </button>
          </div>
        )}
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="form-control">
            <label className="label">
              <span className="label-text font-medium">Workspace Name</span>
            </label>
            <input 
              type="text" 
              name="name"
              placeholder="Enter workspace name" 
              className="input input-bordered w-full" 
              required 
              autoFocus
            />
          </div>
          
          <div className="form-control">
            <label className="label">
              <span className="label-text font-medium">Description (optional)</span>
            </label>
            <textarea 
              name="description"
              className="textarea textarea-bordered w-full h-24 resize-none" 
              placeholder="Enter workspace description"
            ></textarea>
          </div>
          
          <div className="flex justify-end gap-3 pt-4">
            <button 
              type="button" 
              className="btn btn-outline"
              onClick={handleClose}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            {isFreeUser ? (
              <button 
                type="button"
                className="btn btn-primary"
                onClick={() => window.location.href = '/billing'}
              >
                <CreditCard className="w-4 h-4 mr-2" />
                Upgrade Plan
              </button>
            ) : (
              <button 
                type="submit" 
                className="btn btn-primary"
                disabled={isSubmitting}
              >
                {isSubmitting ? (
                  <>
                    <span className="loading loading-spinner loading-sm"></span>
                    Creating...
                  </>
                ) : (
                  'Create Workspace'
                )}
              </button>
            )}
          </div>
        </form>
      </div>
      
      <div className="modal-backdrop" onClick={handleClose}></div>
    </div>
  )
} 