import {createBrowserRouter, Navigate, RouterProvider} from 'react-router-dom'
import Login from './pages/login'
import Notes from './pages/notes'
import { useSession } from './hooks/useSession'

export default function App() {
  const { token, login, logout } = useSession()
  const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
    return token ? <>{children}</> : <Navigate to="/login" />
  }

 const router = createBrowserRouter([
  {
    path: '/login',
    element: token ? <Navigate to="/notes" /> : <Login onLogin={login} />
  },
  {
    path: '/signup',
    element: <></>
  },
  {
    path: '/notes',
    element: (
      <ProtectedRoute>
        <Notes onLogout={logout} />
      </ProtectedRoute>
    )
  },
  {
    path: '*',
    element: <Navigate to={token ? "/notes" : "/login"} />
  }
 ])

 return <RouterProvider router={router} />
}