import React, {useEffect} from "react"
import {Page} from "../../companents/page/Page";
import {
  Container,
  List,
  Typography,
  LinearProgress,
  Paper, Box,
} from "@material-ui/core";
import {useNetworks} from "../../hooks/api/useNetworks";
import {NetworkIListItem} from "./NetworkIListItem";


export const AppPage = () => {
  const {networks, loadNetworks, fetching} = useNetworks();

  useEffect(() => {
    loadNetworks()
  }, [loadNetworks])

  return (
    <Page>
      {fetching && <LinearProgress color='secondary'/>}
      <Container maxWidth='md'>
        <Typography color='textPrimary' variant='h1'>Networks</Typography>
        <Box mt={3}>
          <Paper elevation={2}>
            {networks && <List>
              {networks?.map((n) => (
                <NetworkIListItem key={'network-'+n.name} network={n}/>
              ))}
            </List>}
          </Paper>
        </Box>
      </Container>
    </Page>
  )
}
