import React, { useState } from 'react'
import { X, AlertCircle } from 'react-feather'

export const AddUserModal = ({ isOpen, onClose, onSubmit }) => {
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)

  if (!isOpen) return null

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setIsSubmitting(true)

    const formData = new FormData(e.target)
    const username = formData.get('username')
    const role = formData.get('role')
    
    if (username && role) {
      try {
        const success = await onSubmit(username, role)
        if (success) {
          // Success - modal will be closed by parent
          setError('')
        } else {
          setError('Failed to add user. Please check the email address and try again.')
        }
      } catch (err) {
        setError('Failed to add user. Please try again.')
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

        <h3 className="font-bold text-lg mb-4">Add User to Workspace</h3>
        
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
              <span className="label-text font-medium">Email Address</span>
            </label>
            <input 
              type="email" 
              name="username"
              placeholder="Enter user's email address" 
              className="input input-bordered w-full" 
              required 
              autoFocus
            />
          </div>
          
          <div className="form-control">
            <label className="label">
              <span className="label-text font-medium">Role</span>
            </label>
            <select 
              name="role"
              className="select select-bordered w-full"
              defaultValue="member"
            >
              <option value="member">Member</option>
              <option value="admin">Admin</option>
              <option value="super_admin">Super Admin</option>
            </select>
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
            <button 
              type="submit" 
              className="btn btn-primary"
              disabled={isSubmitting}
            >
              {isSubmitting ? (
                <>
                  <span className="loading loading-spinner loading-sm"></span>
                  Adding...
                </>
              ) : (
                'Add User'
              )}
            </button>
          </div>
        </form>
      </div>
      
      <div className="modal-backdrop" onClick={handleClose}></div>
    </div>
  )
} 