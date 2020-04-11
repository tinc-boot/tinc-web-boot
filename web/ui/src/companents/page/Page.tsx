import React, {ReactNode} from "react"
import {WebBootAppBar} from "../header/WebBootAppBar";
import {Box} from "@material-ui/core";
import {SWAlert} from "../alerts/sw/SWAlert";

type P = {
  children?: ReactNode
}

export const Page = ({children}: P) => {

  return (
    <>
      <WebBootAppBar />
      <Box mt={8}>
        <SWAlert />
        {children}
      </Box>
    </>
  )
}
