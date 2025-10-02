import { useEffect, useState } from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { Flex, ProgressCircle, Text } from '@adobe/react-spectrum';
import { useAuth } from './AuthContext';
export function ProtectedRoute({ children }) {
    const { isAuthenticated, isLoading: authLoading, checkAuth } = useAuth();
    const [isValidating, setIsValidating] = useState(false);
    const location = useLocation();
    useEffect(() => {
        const validateAuth = async () => {
            if (!authLoading && isAuthenticated) {
                setIsValidating(true);
                await checkAuth();
                setIsValidating(false);
            }
        };
        validateAuth();
    }, [isAuthenticated, authLoading, checkAuth]);
    if (authLoading || isValidating) {
        return (<Flex direction="column" alignItems="center" justifyContent="center" height="100vh" gap="size-200">
        <ProgressCircle aria-label="Validating session..." isIndeterminate/>
        <Text>Validating session...</Text>
      </Flex>);
    }
    if (!isAuthenticated) {
        // Redirect to login with return URL
        const returnTo = encodeURIComponent(location.pathname + location.search);
        return <Navigate to={`/login?returnTo=${returnTo}`} replace/>;
    }
    return children ?? <Outlet />;
}
//# sourceMappingURL=ProtectedRoute.jsx.map