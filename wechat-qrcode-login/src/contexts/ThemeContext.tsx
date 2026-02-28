import * as React from "react";

type Theme = "light" | "dark" | "system";

interface ThemeContextType {
  theme: Theme;
  resolvedTheme: "light" | "dark";
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
}

const ThemeContext = React.createContext<ThemeContextType | undefined>(undefined);

const THEME_STORAGE_KEY = "wechat-login-theme";

interface ThemeProviderProps {
  children: React.ReactNode;
  defaultTheme?: Theme;
  switchable?: boolean;
}

// Helper to get initial theme from storage or default
function getInitialTheme(defaultTheme: Theme): Theme {
  if (typeof window === "undefined") return defaultTheme;

  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored && (stored === "light" || stored === "dark" || stored === "system")) {
      return stored as Theme;
    }
  } catch {
    // localStorage might not be available
  }

  return defaultTheme;
}

// Helper to resolve system theme
function resolveSystemTheme(): "light" | "dark" {
  if (typeof window === "undefined") return "light";

  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

// Helper to get resolved theme
function getResolvedTheme(theme: Theme): "light" | "dark" {
  return theme === "system" ? resolveSystemTheme() : theme;
}

export function ThemeProvider({
  children,
  defaultTheme = "system",
  switchable = false,
}: ThemeProviderProps) {
  const [theme, setThemeState] = React.useState<Theme>(() => getInitialTheme(defaultTheme));
  const [resolvedTheme, setResolvedTheme] = React.useState<"light" | "dark">(() =>
    getResolvedTheme(getInitialTheme(defaultTheme))
  );

  // Update DOM and resolved theme when theme changes
  React.useEffect(() => {
    const root = document.documentElement;

    // Remove old theme classes
    root.classList.remove("light", "dark");

    // Add new theme class
    const newResolvedTheme = getResolvedTheme(theme);
    root.classList.add(newResolvedTheme);
    setResolvedTheme(newResolvedTheme);

    // Listen for system theme changes if using system theme
    if (theme === "system") {
      const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");

      const handleChange = () => {
        const newResolved = resolveSystemTheme();
        root.classList.remove("light", "dark");
        root.classList.add(newResolved);
        setResolvedTheme(newResolved);
      };

      // Modern browsers
      mediaQuery.addEventListener("change", handleChange);

      return () => {
        mediaQuery.removeEventListener("change", handleChange);
      };
    }
  }, [theme]);

  // Set theme and persist to localStorage
  const setTheme = React.useCallback(
    (newTheme: Theme) => {
      if (!switchable) return;

      setThemeState(newTheme);

      try {
        localStorage.setItem(THEME_STORAGE_KEY, newTheme);
      } catch {
        // localStorage might not be available
      }
    },
    [switchable]
  );

  // Toggle between light and dark
  const toggleTheme = React.useCallback(() => {
    if (!switchable) return;

    const newTheme: Theme = theme === "light" ? "dark" : "light";
    setThemeState(newTheme);

    try {
      localStorage.setItem(THEME_STORAGE_KEY, newTheme);
    } catch {
      // localStorage might not be available
    }
  }, [switchable, theme]);

  return (
    <ThemeContext.Provider value={{ theme, resolvedTheme, setTheme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = React.useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  return context;
}
