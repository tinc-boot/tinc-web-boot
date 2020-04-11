import {useStateSelector} from "../system/useStateSelector";
import {useCallback} from "react";
import {useApi} from "./useApi";
import {dispatcher} from "../../store";
import {useFetching} from "../system/useFetching";


export function useNetworks() {
  const {api} = useApi();
  const networks = useStateSelector( s => s.networks.list),
    {fetching, withFetching} = useFetching();

  const loadNetworks = useCallback(async () => {
    try {
      const networks = await withFetching(api.networks())
      dispatcher.networks.setList(networks)
    } finally {
    }
  }, [api, withFetching]);

  return {networks, loadNetworks, fetching}
}
