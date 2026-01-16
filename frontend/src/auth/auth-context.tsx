// src/auth/AuthContext.tsx
import { getZitadel } from "@/lib/auth"
import { User } from "oidc-client-ts"
import React, { createContext, useContext, useEffect, useState } from "react"

interface AuthContextType {
  user: User | null
  isLoading: boolean
  login: () => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const zitadel = getZitadel()
    // Check for existing user session on mount
    zitadel.userManager.getUser().then((loadedUser) => {
      if (loadedUser && !loadedUser.expired) {
        setUser(loadedUser)
      }
      setIsLoading(false)
    })

    // Subscribe to user changes (silent renew, logout, etc.)
    const handleUserLoaded = (u: User) => setUser(u)
    const handleUserUnloaded = () => setUser(null)

    zitadel.userManager.events.addUserLoaded(handleUserLoaded)
    zitadel.userManager.events.addUserUnloaded(handleUserUnloaded)

    return () => {
      zitadel.userManager.events.removeUserLoaded(handleUserLoaded)
      zitadel.userManager.events.removeUserUnloaded(handleUserUnloaded)
    }
  }, [])

  const login = () => getZitadel().authorize()
  const logout = () => getZitadel().signout()

  if (isLoading) return <div>Loading Auth...</div> // Or a proper skeleton loader

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

// Custom Hook for easy usage
export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) throw new Error("useAuth must be used within an AuthProvider")
  return context
}