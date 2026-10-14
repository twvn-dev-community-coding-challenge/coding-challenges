import { create } from 'zustand'

const useStore = create((set) => ({
  message: 'Loading from Zustand...',
  setMessage: (newMessage) => set({ message: newMessage }),
  error: null,
  setError: (newError) => set({ error: newError }),
}))

export default useStore
