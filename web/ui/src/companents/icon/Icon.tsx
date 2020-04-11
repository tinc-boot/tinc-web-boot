import React, {useMemo} from "react";
import {FontAwesomeIcon, FontAwesomeIconProps} from "@fortawesome/react-fontawesome";
import {fa} from "../../bootstrap/fa";
import {styled} from "@material-ui/core";

export type IconType = keyof typeof fa

type P = {
  icon: IconType
} & Omit<FontAwesomeIconProps, 'icon'>

const StyledFontAwesomeIcon = styled(FontAwesomeIcon)({
  transitionProperty: 'color',
  transitionDuration: '.3s',
  transitionTimingFunction: 'ease-in-out'
})

export const Icon = (p: P) => {
  const {icon, ...restProps} = p;
  const i = useMemo(() => fa[icon], [icon]);

  return (
    <StyledFontAwesomeIcon icon={i} {...restProps} />
  )
};
