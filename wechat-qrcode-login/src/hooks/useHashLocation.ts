/**
 * Custom hash location hook for wouter
 * Properly handles query parameters in hash-based routing
 *
 * Fixes issue where URLs like /#/oauth/callback?code=xxx would not match routes
 */
import { useEffect, useState } from 'react';

export function useHashLocation(): [string, (path: string) => void] {
  const [location, setLocation] = useState(() => {
    // Extract path from hash, removing query parameters for route matching
    const hash = window.location.hash.slice(1) || '/';
    // Split by ? to get only the path part
    const path = hash.split('?')[0];
    console.log('[useHashLocation] Initial hash:', hash, '-> path:', path);
    return path;
  });

  useEffect(() => {
    const handleHashChange = () => {
      const hash = window.location.hash.slice(1) || '/';
      // Split by ? to get only the path part for route matching
      const path = hash.split('?')[0];
      console.log('[useHashLocation] Hash changed:', hash, '-> path:', path);
      setLocation(path);
    };

    // Listen for hash changes
    window.addEventListener('hashchange', handleHashChange);

    return () => {
      window.removeEventListener('hashchange', handleHashChange);
    };
  }, []);

  const navigate = (path: string) => {
    console.log('[useHashLocation] Navigate to:', path);
    window.location.hash = path;
  };

  return [location, navigate];
}
