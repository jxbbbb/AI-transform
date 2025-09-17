import {ErrorMsg} from "./axios"
import request from "./axios"
import type {AxiosProgressEvent,AxiosResponse,GenericAbortSignal,AxiosError} from 'axios'
export interface HttpParams{
    url: string
    data?: any
    method?: string
    headers?: any
    onDownloadProgress?: (progressEvent: AxiosProgressEvent) => void
    onUploadProgress?:(progressEvent: AxiosProgressEvent) => void
    signal?: GenericAbortSignal
    beforeRequest?: () => void
    afterRequest?: () => void
  }
  export interface Response<T=any> {
    data?: T
    status: number 
    message?: string 
  }

  export function http<T = any>(
    {url,data,method,headers,onDownloadProgress,onUploadProgress,signal,beforeRequest,afterRequest}: HttpParams
    ){
    const successHandler = (res:AxiosResponse<T>) => {
        return Promise.resolve({data:res.data,status:res.status})
    }
    const failHandler = (err: Error) => {
        const axiosErr = err as AxiosError
        let message = axiosErr.message
        if (axiosErr.response?.data) {
          let errorMsg = axiosErr.response.data as ErrorMsg
          if (errorMsg.error) {
              message = errorMsg.error
          }
      }
      return Promise.reject({status:axiosErr.response?.status,message:message})
    }
    beforeRequest?.()
    method = method||'GET'
    const params = Object.assign(typeof data === 'function'?data():data??{},{})

    switch (method) {
      case 'GET':
        return request.get(url, { params, signal, onDownloadProgress }).then(successHandler, failHandler).finally(afterRequest)
      case 'POST':
        return request.post(url, params, { headers, signal, onDownloadProgress,onUploadProgress }).then(successHandler, failHandler).finally(afterRequest)
      case 'PUT':
        return request.put(url, params, { headers, signal, onDownloadProgress,onUploadProgress }).then(successHandler, failHandler).finally(afterRequest)
      default:
        return request.post(url, params, { headers, signal, onDownloadProgress,onUploadProgress }).then(successHandler, failHandler).finally(afterRequest)
    }

    return method === 'GET'
    ? request.get(url, { params, signal, onDownloadProgress }).then(successHandler, failHandler).finally(afterRequest)
    : request.post(url, params, { headers, signal, onDownloadProgress,onUploadProgress }).then(successHandler, failHandler).finally(afterRequest)
  }

  export function get<T = any>(
    { url, data, method = 'GET', headers, onDownloadProgress, signal, beforeRequest, afterRequest }: HttpParams,
  ): Promise<Response<T>> {
    return http<T>({
      url,
      method,
      data,
      headers,
      onDownloadProgress,
      signal,
      beforeRequest,
      afterRequest,
    })
  }
  
  export function post<T = any>(
    { url, data, method = 'POST', headers, onDownloadProgress,onUploadProgress, signal, beforeRequest, afterRequest }: HttpParams,
  ): Promise<Response<T>> {
    return http<T>({
      url,
      method,
      data,
      headers,
      onDownloadProgress,
      onUploadProgress,
      signal,
      beforeRequest,
      afterRequest,
    })
  }
  
  export default post