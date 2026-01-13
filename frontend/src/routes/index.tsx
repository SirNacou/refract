import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({ component: HomePage })

function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900">
      <div className="max-w-4xl mx-auto px-6 py-20">
        <h1 className="text-4xl font-bold text-white mb-4">
          Authenticated URL Shortener
        </h1>
        <p className="text-gray-400 text-lg">
          Welcome to Refract - A distributed URL shortening platform
        </p>
      </div>
    </div>
  )
}
