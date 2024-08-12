import { verifyJWT } from "../hoc/JWT";

export default function useIsAuthenticated() {
    return verifyJWT();
}