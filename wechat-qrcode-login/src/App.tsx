import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Router, Route, Switch } from "wouter";
import { useHashLocation } from "@/hooks/useHashLocation";
import ErrorBoundary from "@/components/ErrorBoundary";
import { ThemeProvider } from "@/contexts/ThemeContext";
import { ThemeToggle } from "@/components/ThemeToggle";
import Landing from "@/pages/Landing";
import Home from "@/pages/Home";
import Login from "@/pages/Login";
import OAuthAuthorize from "@/pages/OAuthAuthorize";
import OAuthCallback from "@/pages/OAuthCallback";
import NotFound from "@/pages/NotFound";

// Use hash-based routing with custom hook that properly handles query parameters
function AppRouter() {
  return (
    <Router hook={useHashLocation}>
      <Switch>
        {/* Default route shows Landing page */}
        <Route path="/" component={Landing} />
        <Route path="/login" component={Login} />
        <Route path="/home" component={Home} />
        <Route path="/oauth/authorize" component={OAuthAuthorize} />
        <Route path="/oauth/callback" component={OAuthCallback} />
        <Route component={NotFound} />
      </Switch>
    </Router>
  );
}

function App() {
  return (
    <ErrorBoundary>
      {/* Use system theme as default, with user choice persisted */}
      <ThemeProvider defaultTheme="system" switchable={true}>
        <TooltipProvider>
          <Toaster position="top-center" />
          <ThemeToggle />
          <AppRouter />
        </TooltipProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
