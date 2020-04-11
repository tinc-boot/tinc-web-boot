import {useSelector} from "react-redux";
import {AppState} from "../../store";


export function useStateSelector<T = unknown>(
  selector: (state: AppState) => T,
  equalityFn?: (left: T, right: T) => boolean
): T {
  return useSelector(selector, equalityFn)
}
