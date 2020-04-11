import React from "react"
import {BrowserRouter, Switch, Route, Redirect} from "react-router-dom";
import {AppPage} from "../pages/app/AppPage";


export const Routing = () => {
  return (
    <BrowserRouter>
      <Switch>
        <Route path='/app'>
          <AppPage />
        </Route>
        <Route exact path='/'>
          <Redirect to='/app' />
        </Route>
      </Switch>
    </BrowserRouter>
  )
}
