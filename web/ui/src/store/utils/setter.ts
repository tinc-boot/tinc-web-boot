import {PayloadAction} from "@reduxjs/toolkit";

export const setter = <S, F extends keyof S>(initialState: S, f: F, prepare?: (data: S[F]) => S[F]) => (s: S, a?: PayloadAction<S[F]>): void => {
  s[f] = a ? a.payload : initialState[f];
  if (prepare) {
    s[f] = prepare(s[f])
  }
};
