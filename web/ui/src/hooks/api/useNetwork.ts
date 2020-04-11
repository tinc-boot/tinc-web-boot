import {Network} from "../../api/api";
import {useCallback} from "react";
import {useApi} from "./useApi";
import {useFetching} from "../system/useFetching";
import {dispatcher} from "../../store";


export function useNetwork(network: Network) {
  const {api} = useApi(),
    {fetching, withFetching} = useFetching();

  const start = useCallback(async () => {
    const res = await withFetching(api.start(network.name));
    dispatcher.networks.add(res)
  }, [api, network.name, withFetching]);

  const stop = useCallback(async () => {
    const res = await withFetching(api.stop(network.name));
    dispatcher.networks.add(res)
  }, [api, network.name, withFetching])

  return {start, fetching, stop}
}
