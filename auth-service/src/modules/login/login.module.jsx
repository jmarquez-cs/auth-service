import React, {Component} from 'react';
import LoginForm from "./forms/login-form.container";

class Login extends Component {
    render() {
        return (
            <>
                <LoginForm onForgotPasswordOpen={this.handleForgotPasswordOpen} onForgotPasswordClose={this.handleForgotPasswordClose}/>
            </>
        );
    }

}

export default Login;