import {useCallback, useState} from "react";


export const useFetching = () => {
  const [fetching, setFetching] = useState(false);

  const withFetching = useCallback( <T>(task: Promise<T>): Promise<T> => {
    setFetching(true);
    task.finally(() => {
      setTimeout(() => setFetching(false), 300)
    });
    return task;

  }, []);

  return {fetching, withFetching}
}
