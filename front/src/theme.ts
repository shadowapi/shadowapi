import type { ThemeConfig } from 'antd';

// Color palette based on: https://coolors.co/palette/000000-14213d-fca311-e5e5e5-ffffff
export const colors = {
  black: '#000000',
  oxfordBlue: '#14213d',
  orange: '#fca311',
  lightGray: '#e5e5e5',
  white: '#ffffff',
};

// Derived colors for UI elements
export const uiColors = {
  // Header/Navigation
  headerBg: colors.oxfordBlue,
  headerBorder: colors.black,

  // Logo
  logoBg: colors.orange,
  logoText: colors.black,

  // Menu
  menuText: colors.lightGray,
  menuTextHover: colors.white,
  menuTextActive: colors.white,
  menuItemHoverBg: 'rgba(252, 163, 17, 0.15)', // orange with low opacity
  menuItemActiveBg: 'rgba(252, 163, 17, 0.25)', // orange with higher opacity

  // Footer
  footerBg: colors.lightGray,
  footerText: colors.oxfordBlue,
  footerBorder: '#d0d0d0',

  // Content
  contentBg: colors.white,

  // Accent
  primary: colors.orange,
  primaryHover: '#e5940f', // slightly darker orange
};

// Ant Design theme configuration
export const theme: ThemeConfig = {
  token: {
    colorPrimary: colors.orange,
    colorLink: colors.orange,
    colorLinkHover: uiColors.primaryHover,
    colorPrimaryBgHover: 'transparent', // Remove beige hover background on links
    colorBgContainer: colors.white,
    colorBgLayout: colors.lightGray,
    borderRadius: 6,
  },
  components: {
    Button: {
      colorPrimary: colors.orange,
      algorithm: true,
    },
    Menu: {
      darkItemBg: 'transparent',
      darkItemColor: uiColors.menuText,
      darkItemHoverColor: uiColors.menuTextHover,
      darkItemHoverBg: uiColors.menuItemHoverBg,
      darkItemSelectedBg: uiColors.menuItemActiveBg,
      darkItemSelectedColor: uiColors.menuTextActive,
      horizontalItemBorderRadius: 6,
      horizontalItemHoverBg: uiColors.menuItemHoverBg,
      horizontalItemSelectedBg: uiColors.menuItemActiveBg,
      // Light menu (sidebar) - remove active/hover background for submenu titles
      itemActiveBg: 'transparent',
      subMenuItemBg: 'transparent',
    },
  },
};

export default theme;
