import {IconButton, ListItem, ListItemIcon, ListItemSecondaryAction, ListItemText, useTheme} from "@material-ui/core";
import {Icon, IconType} from "../../companents/icon/Icon";
import React, {useMemo} from "react";
import {Network} from "../../api/api";
import {useNetwork} from "../../hooks/api/useNetwork";

type P = {
  network: Network
}

export const NetworkIListItem = (p: P) => {
  const theme = useTheme();
  const {start, stop, fetching} = useNetwork(p.network);

  const {icon, onAction} = useMemo(() => ({
    onAction: p.network.running ? stop : start,
    icon: (p.network.running ? 'faPause' : 'faPlay' as IconType)
  }), [p.network.running, start, stop]);

  const iconColor = useMemo(() =>
    fetching
      ? theme.palette.success.dark
      : p.network.running
        ? theme.palette.success.main
      : theme.palette.background.default,
    [fetching, p.network.running, theme.palette.background.default, theme.palette.success.dark, theme.palette.success.main]);

  return (
    <ListItem disabled={fetching}>
      <ListItemIcon>
        <Icon icon='faNetworkWired' color={iconColor} />
      </ListItemIcon>
      <ListItemText primary={p.network.name} />
      <ListItemSecondaryAction>
        <IconButton onClick={onAction} disabled={fetching}>
          <Icon icon={fetching ? 'faSpinner' : icon} spin={fetching} />
        </IconButton>
      </ListItemSecondaryAction>
    </ListItem>
  )
}
