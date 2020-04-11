import { library } from '@fortawesome/fontawesome-svg-core'
import {faNetworkWired, faPause, faPlay, faSpinner} from '@fortawesome/free-solid-svg-icons'
import _ from 'lodash'
// eslint-disable-next-line @typescript-eslint/no-unused-vars
import {IconDefinition} from "@fortawesome/fontawesome-common-types";

// IconDefinition
//const i: IconDefinition = faPlay

export const fa = {
  faSpinner,
  faNetworkWired,
  faPlay,
  faPause
};

export const faInit = () => {
  library.add(..._.values(fa))
};
