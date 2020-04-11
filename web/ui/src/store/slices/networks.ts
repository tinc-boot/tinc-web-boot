import {Network} from "../../api/api";
import {createSlice, PayloadAction} from "@reduxjs/toolkit";
import {setter} from "../utils/setter";
import _ from "lodash";


export interface NetworksState {
  list?: Network[]
}

const initialState: NetworksState = {}

const sort = (d?: Network[]): Network[] => _.sortBy(d, 'name')

export const networksSlice = createSlice({
  name: 'networks',
  initialState,
  reducers: {
    setList: setter(initialState, 'list', sort),
    add: (s: NetworksState, a: PayloadAction<Network>) => {
      const list = s.list || [];
      _.remove(list, n => n.name === a.payload.name)
      s.list = sort([...list, a.payload])
    }
  }
})
