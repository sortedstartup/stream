import { toast, ToastOptions, ToastContent, Id } from 'react-toastify'

// Default toast configuration
const defaultConfig: ToastOptions = {
  position: "top-center",
  autoClose: 5000,
  hideProgressBar: false,
  closeOnClick: true,
  pauseOnHover: true,
  draggable: true,
  theme: "light",
}

// Success toast
export const showSuccessToast = (message: ToastContent, customConfig: ToastOptions = {}): Id => {
  return toast.success(message, {
    ...defaultConfig,
    ...customConfig
  })
}

// Error toast
export const showErrorToast = (message: ToastContent, customConfig: ToastOptions = {}): Id => {
  return toast.error(message, {
    ...defaultConfig,
    autoClose: 7000, // Errors stay longer
    ...customConfig
  })
}

// Warning toast
export const showWarningToast = (message: ToastContent, customConfig: ToastOptions = {}): Id => {
  return toast.warn(message, {
    ...defaultConfig,
    autoClose: 6000,
    ...customConfig
  })
}

// Info toast
export const showInfoToast = (message: ToastContent, customConfig: ToastOptions = {}): Id => {
  return toast.info(message, {
    ...defaultConfig,
    autoClose: 4000,
    ...customConfig
  })
}