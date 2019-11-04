export const authenticationReducer = (state = {}, action) => {
    switch(action.type) {
        case 'LOGIN_REQUEST': {
            return state = { loginUserSuccess: action.payload }
        } 
        case 'LOGIN_USER_SUCCESS': {
            return state = { loginUserSuccess: action.payload }
        }
        case 'LOGIN_USER_FAILED': {
            return state = { loginUserFailed: action.payload }
        }
        case 'LOGOUT_REQUEST': {
            return state = { loggingOut: action.payload }
        }
        default:
            return state
    }
};