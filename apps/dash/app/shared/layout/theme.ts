import { type ThemeConfig, theme } from "antd";

export type ThemeMode = "light" | "dark";

const palette = {
  light: {
    primary: "#b85630",
    primaryHover: "#a94d29",
    primaryActive: "#94421f",
    onPrimary: "#ffffff",
    tint: "rgba(194, 96, 58, 0.10)",
    tintHover: "rgba(194, 96, 58, 0.16)",
    bg: "#faf6f0",
    text: "#2b211b",
    border: "rgba(43, 33, 27, 0.12)",
    borderSoft: "rgba(43, 33, 27, 0.07)",
    glass: "rgba(255, 255, 255, 0.55)",
    glassStrong: "rgba(255, 255, 255, 0.72)",
    success: "#4f7f52",
    warning: "#b7791f",
    danger: "#b83a4b",
    info: "#4a6fa5",
  },
  dark: {
    primary: "#e08a66",
    primaryHover: "#eba283",
    primaryActive: "#c9714d",
    onPrimary: "#2a140b",
    tint: "rgba(224, 138, 102, 0.16)",
    tintHover: "rgba(224, 138, 102, 0.24)",
    bg: "#1c1713",
    text: "#f3eae2",
    border: "rgba(243, 234, 226, 0.14)",
    borderSoft: "rgba(243, 234, 226, 0.08)",
    glass: "rgba(255, 255, 255, 0.06)",
    glassStrong: "rgba(255, 255, 255, 0.10)",
    success: "#7fb382",
    warning: "#e0b04f",
    danger: "#e0707f",
    info: "#82a6d6",
  },
} as const;

export function getAntdTheme(mode: ThemeMode): ThemeConfig {
  const p = palette[mode];

  return {
    algorithm: [
      mode === "dark" ? theme.darkAlgorithm : theme.defaultAlgorithm,
      theme.compactAlgorithm,
    ],
    token: {
      colorPrimary: p.primary,
      colorPrimaryHover: p.primaryHover,
      colorPrimaryActive: p.primaryActive,
      colorPrimaryBg: p.tint,
      colorPrimaryBgHover: p.tintHover,
      colorLink: p.primary,
      colorLinkHover: p.primaryHover,
      colorLinkActive: p.primaryActive,

      colorSuccess: p.success,
      colorWarning: p.warning,
      colorError: p.danger,
      colorInfo: p.info,

      colorBgBase: p.bg,
      colorTextBase: p.text,
      colorBgLayout: "transparent",
      colorBgContainer: p.glass,
      colorBgElevated: p.glassStrong,
      colorBorder: p.border,
      colorBorderSecondary: p.borderSoft,

      borderRadius: 10,
      borderRadiusSM: 6,
      borderRadiusLG: 14,

      fontFamily:
        '"Inter", ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif',
      fontSize: 15,
      fontSizeHeading1: 28,
      fontSizeHeading2: 22,
      fontSizeHeading3: 18,
      fontSizeHeading4: 16,
      fontSizeHeading5: 15,

      wireframe: false,
      boxShadow: "none",
      boxShadowSecondary: "0 4px 24px rgba(0, 0, 0, 0.08)",
    },
    components: {
      Button: {
        primaryColor: p.onPrimary,
        primaryShadow: "none",
        defaultShadow: "none",
        dangerShadow: "none",
        fontWeight: 500,
      },
      Layout: {
        bodyBg: "transparent",
        headerBg: "transparent",
        siderBg: "transparent",
        triggerBg: "transparent",
        headerHeight: 52,
        headerPadding: "0 16px",
      },
      Menu: {
        itemBg: "transparent",
        subMenuItemBg: "transparent",
        itemHoverBg: p.tint,
        itemSelectedBg: p.tint,
        itemSelectedColor: p.primary,
        itemBorderRadius: 8,
        activeBarBorderWidth: 0,
      },
      Table: {
        headerBg: "transparent",
        rowHoverBg: p.tint,
        borderColor: p.borderSoft,
      },
      Card: {
        headerHeight: 44,
        bodyPadding: 16,
        headerPadding: 16,
      },
      Modal: {
        contentBg: p.glassStrong,
        headerBg: "transparent",
        footerBg: "transparent",
      },
    },
  };
}
