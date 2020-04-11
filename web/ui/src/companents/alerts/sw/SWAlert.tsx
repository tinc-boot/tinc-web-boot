import React, {useCallback} from "react"
import {Alert} from "@material-ui/lab";
import {useStateSelector} from "../../../hooks/system/useStateSelector";
import {Button} from "@material-ui/core";
import {SWStatus} from "../../../store/slices/system";


export const SWAlert = () => {
  const status = useStateSelector( s => s.system.status),
    sw = useStateSelector(s => s.system.sw);

  const onClick = useCallback(() => {
    if (sw) {
      sw.postMessage({
        type: 'SKIP_WAITING'
      });
      setTimeout(() => {
        window.location.reload();
      }, 500);
    } else {
      console.warn('sw is required started!')
    }
  }, [sw]);

  if (status !== SWStatus.WAITING) {
    return <></>;
  }

  return (
    <Alert severity='warning' action={
      <Button color="inherit" size="small" onClick={onClick}>
        FORCE UPDATE
      </Button>
    }>
      <div>TincWebBoot site has been updated, please press the button to update this session. Please reload all other tabs where crix.io is open</div>
    </Alert>
  )
}
