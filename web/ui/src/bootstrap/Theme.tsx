import React, {useMemo} from "react"
import {createMuiTheme, useMediaQuery, ThemeProvider, ThemeProviderProps, responsiveFontSizes} from "@material-ui/core";

type P = Omit<ThemeProviderProps, 'theme'>

export const Theme = ({children, ...rest}: P) => {
  const isDarkMode = useMediaQuery('(prefers-color-scheme: dark)');

  const theme = useMemo(() => responsiveFontSizes(createMuiTheme({
    palette: {
      type: isDarkMode ? 'dark' : 'light',
    },
    typography: {
      h1: {fontSize: '4rem'},
      h2: {fontSize: '3.5rem'},
    }
  })), [isDarkMode]);

  return (
    <ThemeProvider theme={theme} children={children} {...rest} />
  )
}
