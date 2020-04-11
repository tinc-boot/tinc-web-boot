import {useCallback, useEffect} from "react";
import {TincWeb} from "../../api/api";

let api: TincWeb = new TincWeb()

export function useApi() {
  const createApi = useCallback(() => {
    api = new TincWeb()
  }, []);

  useEffect(() => {
    if (!api) createApi()
  }, [createApi])

  return {api, createApi}
}
