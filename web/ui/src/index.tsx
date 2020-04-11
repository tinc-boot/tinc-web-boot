import './index.css'

import React from 'react';
import ReactDOM from 'react-dom';
import * as serviceWorker from './serviceWorker';
import {dispatcher} from "./store";
import {SWStatus} from "./store/slices/system";
import {App} from "./bootstrap/App";
import {faInit} from "./bootstrap/fa";

faInit()

ReactDOM.render(
  <React.StrictMode>
    <App/>
  </React.StrictMode>,
  document.getElementById('root')
);

serviceWorker.register({
  onRegistration: reg => {
    if (reg.waiting || reg.installing) {
      dispatcher.system.setStatus(SWStatus.WAITING)
    }
    if (reg.active) {
      dispatcher.system.setSW(reg.active)
    }
  },
  onUpdate: reg => {
    if (reg.waiting || reg.installing) {
      dispatcher.system.setStatus(SWStatus.WAITING)
    }
    if (reg.active) {
      dispatcher.system.setSW(reg.active)
    }
  },
  onSuccess: reg => {
    dispatcher.system.setStatus(SWStatus.ACTIVE);
    if (reg.active) {
      dispatcher.system.setSW(reg.active)
    }
  }
});
