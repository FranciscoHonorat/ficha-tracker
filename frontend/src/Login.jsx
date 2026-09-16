import {useEffect, useState} from 'react';
import {CreateAccount, HasAccount, Login as LoginCall} from "../wailsjs/go/main/App";

function Login({onAuthenticated}) {
    const [checking, setChecking] = useState(true);
    const [hasAccount, setHasAccount] = useState(false);
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [error, setError] = useState('');
    const [submitting, setSubmitting] = useState(false);

    useEffect(() => {
        HasAccount()
            .then(setHasAccount)
            .catch((err) => setError(String(err)))
            .finally(() => setChecking(false));
    }, []);

    const submit = (e) => {
        e.preventDefault();
        setError('');

        if (!hasAccount && password !== confirmPassword) {
            setError('As senhas não coincidem.');
            return;
        }

        setSubmitting(true);
        const action = hasAccount
            ? LoginCall(username.trim(), password)
            : CreateAccount(username.trim(), password);

        action
            .then(() => onAuthenticated())
            .catch((err) => setError(String(err)))
            .finally(() => setSubmitting(false));
    };

    if (checking) {
        return <div id="App" className="login-screen"/>;
    }

    return (
        <div id="App" className="login-screen">
            <div className="card login-card">
                <h1>Ficha Tracker</h1>
                <p className="login-subtitle">
                    {hasAccount ? 'Entrar na conta' : 'Criar a conta local do aplicativo'}
                </p>
                <form onSubmit={submit} className="form">
                    <label>
                        Usuário
                        <input
                            type="text"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            autoComplete="username"
                            autoFocus
                            required
                        />
                    </label>
                    <label>
                        Senha
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            autoComplete={hasAccount ? 'current-password' : 'new-password'}
                            required
                        />
                    </label>
                    {!hasAccount && (
                        <label>
                            Confirmar senha
                            <input
                                type="password"
                                value={confirmPassword}
                                onChange={(e) => setConfirmPassword(e.target.value)}
                                autoComplete="new-password"
                                required
                            />
                        </label>
                    )}
                    <button type="submit" className="btn" disabled={submitting}>
                        {submitting ? 'Aguarde...' : (hasAccount ? 'Entrar' : 'Criar conta')}
                    </button>
                </form>
                {error && <p className="message error">{error}</p>}
            </div>
        </div>
    );
}

export default Login;
