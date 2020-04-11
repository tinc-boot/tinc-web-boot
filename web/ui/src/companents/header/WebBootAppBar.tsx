import React from "react"
import {AppBar, Toolbar, Typography} from "@material-ui/core";



export const WebBootAppBar = () => {

  return (
    <AppBar position='fixed' color='default'>
      <Toolbar>
        <Typography variant='h6'>TincWebBoot</Typography>
      </Toolbar>
    </AppBar>
  )
}
