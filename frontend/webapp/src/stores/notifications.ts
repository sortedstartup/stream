import { atom } from 'nanostores'

// Success message store
export const $successMessage = atom<string | null>(null)

// Helper functions
export const showSuccessMessage = (message: string) => {
  $successMessage.set(message)
  // Auto-clear after 5 seconds
  setTimeout(() => {
    $successMessage.set(null)
  }, 5000)
}

export const clearSuccessMessage = () => {
  $successMessage.set(null)
} 