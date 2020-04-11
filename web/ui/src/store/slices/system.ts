import {createSlice} from "@reduxjs/toolkit";
import {setter} from "../utils/setter";

export enum SWStatus {
  WAITING,
  ACTIVE
}


export interface SystemState {
  status?: SWStatus,
  sw?: ServiceWorker,
}

const initialState: SystemState = {};

export const systemSlice = createSlice({
  name: 'system',
  initialState,
  reducers: {
    setStatus: setter(initialState, 'status'),
    setSW: setter(initialState, 'sw'),
  }
});
