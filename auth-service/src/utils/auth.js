export const isLoggedIn = () => {
    return localStorage.getItem('token');
};

export const loginErrorHandler = (error) => {
    const statusCode = error.response.status;

    switch (statusCode) {
        case 404:
            return 'Invalid email/password combination.';
        case 403:
            return 'Invalid email/password combination.';
        case 500:
            return 'An error occurred while logging in user.';
        default:
            return 'An error occurred while logging in user.';
    }
};