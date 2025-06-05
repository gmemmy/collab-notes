import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { authApi } from '@/services/api'


export function useLogin(onSuccess: (token: string) => void) {
  const navigate = useNavigate()

  return useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      onSuccess(data.token)
      navigate('/notes')
    },
  })
}

export function useSignUp(onSuccess: (token: string) => void) {
  const navigate = useNavigate()

  return useMutation({
    mutationFn: authApi.signUp,
    onSuccess: (data) => {
      onSuccess(data.token)
      navigate('/notes')
    },
  })
} 