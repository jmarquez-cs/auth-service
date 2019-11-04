import React, { Component } from 'react';
import Login from './login/login.module';
import Registration from './registration/registration.module'
import { Redirect, Route, Switch } from 'react-router-dom';
import { isLoggedIn } from '../utils/auth';

class Core extends Component {
    render() {
        return (
            <Switch>
                <Route path={"/register"} component={Registration} />
                <Route path={"/login"}
                    render={(props) => isLoggedIn()
                        ? <Redirect to={{ pathname: "", state: { from: props.location } }} />
                        : <Login {...props} />}
                />
                <Route path={"/"}
                    render={(props) => isLoggedIn()
                        ? console.log("Successful login => <NavBar {...props} />")
                        : <Redirect to={{ pathname: "login", state: { from: props.location } }} />}
                />
            </Switch>
        );
    }
}

export default Core;