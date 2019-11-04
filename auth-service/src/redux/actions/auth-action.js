import axios from 'axios';
import { authHeader } from '../auth-header'
import { history } from '../../history';

const header = {
    'Content-Type': 'application/json',
};

export const registerUser = (data) => {
    return async (dispatch) => {
        dispatch({
            type: 'REGISTER_REQUEST',
            payload: true,
        });

        await axios.post('http://localhost:8000/api/register', data, header)
        .then((response) => {
            dispatch({
                type: 'REGISTER_USER_SUCCESS',
                payload: true,
            });

            localStorage.setItem('token', response.data.token);
            localStorage.setItem('user', JSON.stringify(response.data.user));
        })
        .catch((error) => {
            dispatch({
                type: 'REGISTER_USER_FAILED',
                payload: error,
            })
        });
    };
};

export const loginUser = (data) => {
    return async (dispatch) => {
        dispatch({
            type: 'LOGIN_REQUEST',
            payload: true,
        });
        await axios.post('http://localhost:8000/api/login', data, header)
        .then((response) => {
            localStorage.setItem('token', response.data.token);

            dispatch({
                type: 'LOGIN_USER_SUCCESS',
                payload: response.data,
            });

            if (response.data.token) {
                localStorage.setItem('token', response.data.token);

                history.push('/');
            }
        })
        .catch((error) => {
            dispatch({
                type: 'LOGIN_USER_FAILED',
                payload: error,
            });
        });
    };
};