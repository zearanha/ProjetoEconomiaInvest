import { useCallback, useEffect, useState } from 'react'

export type ResourceState<T> = { data: T | null; error: string; loading: boolean; refresh: () => Promise<void> }
export function useResource<T>(load: () => Promise<T>, deps: unknown[] = [], interval = 60_000): ResourceState<T> {
 const [data,setData]=useState<T|null>(null),[error,setError]=useState(''),[loading,setLoading]=useState(true)
 const refresh=useCallback(async()=>{try{setError('');setData(await load())}catch(e){setError(e instanceof Error?e.message:'Falha ao carregar dados')}finally{setLoading(false)}
 // load is memoized at call sites; deps define when the request changes.
 // eslint-disable-next-line react-hooks/exhaustive-deps
 },deps)
 useEffect(()=>{void refresh();const timer=window.setInterval(()=>void refresh(),interval);return()=>window.clearInterval(timer)},[refresh,interval])
 return {data,error,loading,refresh}
}
