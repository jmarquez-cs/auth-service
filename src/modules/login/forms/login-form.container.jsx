
import React, {Component} from 'react';
import LoginFormView from "./login-form.view";


class LoginForm extends Component {
    constructor(props) {
        super(props);

        this.state = {
            user: {
                email: '',
                password: '',
            },
        };
    }

    render() {
        const {loggingIn} = this.props;
        return (
            <>
            </>
        );
    }
}