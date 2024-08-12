import { Navigate, useLocation } from "react-router-dom";
import useIsAuthenticated from "../../hooks/useIsAuthenticated";

export default function AuthRoute(props: { children: React.ReactNode }) {
    const { children } = props
    const location = useLocation()
    const isLoggedIn = useIsAuthenticated()

    return isLoggedIn ? (
        <>{children}</>
    ) : (
        <Navigate replace={true} to="/login" state={{ from: `${location.pathname}${location.search}`}} />
    )
}