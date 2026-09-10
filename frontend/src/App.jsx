import {useEffect, useState} from 'react';
import './App.css';
import {ListFichas, RegisterFicha, SearchFichasByName} from "../wailsjs/go/main/App";

const emptyForm = {fullName: '', requestType: '', acs: ''};

function formatDateTime(isoString) {
    const date = new Date(isoString);
    if (isNaN(date.getTime())) return isoString;
    return date.toLocaleString('pt-BR');
}

function App() {
    const [form, setForm] = useState(emptyForm);
    const [fichas, setFichas] = useState([]);
    const [query, setQuery] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [submitting, setSubmitting] = useState(false);

    const loadFichas = (name) => {
        const request = name ? SearchFichasByName(name) : ListFichas();
        request.then(setFichas).catch((err) => setError(String(err)));
    };

    useEffect(() => {
        loadFichas('');
    }, []);

    useEffect(() => {
        const timeout = setTimeout(() => loadFichas(query.trim()), 250);
        return () => clearTimeout(timeout);
    }, [query]);

    const updateField = (field) => (e) => {
        setForm((prev) => ({...prev, [field]: e.target.value}));
    };

    const submit = (e) => {
        e.preventDefault();
        setError('');
        setSuccess('');
        setSubmitting(true);

        RegisterFicha(form.fullName.trim(), form.requestType.trim(), form.acs.trim())
            .then((ficha) => {
                setSuccess(`Ficha de ${ficha.fullName} registrada com sucesso.`);
                setForm(emptyForm);
                loadFichas(query.trim());
            })
            .catch((err) => setError(String(err)))
            .finally(() => setSubmitting(false));
    };

    return (
        <div id="App">
            <header className="header">
                <h1>Ficha Tracker</h1>
                <p>Registro de impressão de fichas de saúde</p>
            </header>

            <main className="content">
                <section className="card">
                    <h2>Registrar ficha</h2>
                    <form onSubmit={submit} className="form">
                        <label>
                            Nome completo
                            <input
                                type="text"
                                value={form.fullName}
                                onChange={updateField('fullName')}
                                autoComplete="off"
                                required
                            />
                        </label>
                        <label>
                            Tipo de solicitação
                            <input
                                type="text"
                                value={form.requestType}
                                onChange={updateField('requestType')}
                                placeholder="Ex.: Exame de sangue"
                                autoComplete="off"
                                required
                            />
                        </label>
                        <label>
                            ACS
                            <input
                                type="text"
                                value={form.acs}
                                onChange={updateField('acs')}
                                autoComplete="off"
                                required
                            />
                        </label>
                        <button type="submit" className="btn" disabled={submitting}>
                            {submitting ? 'Registrando...' : 'Registrar'}
                        </button>
                    </form>
                    {error && <p className="message error">{error}</p>}
                    {success && <p className="message success">{success}</p>}
                </section>

                <section className="card">
                    <div className="list-header">
                        <h2>Fichas registradas</h2>
                        <input
                            type="text"
                            className="search"
                            placeholder="Buscar por nome..."
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                        />
                    </div>

                    {fichas.length === 0 ? (
                        <p className="empty">Nenhuma ficha encontrada.</p>
                    ) : (
                        <table className="table">
                            <thead>
                            <tr>
                                <th>Nome</th>
                                <th>Tipo de solicitação</th>
                                <th>ACS</th>
                                <th>Registrado em</th>
                            </tr>
                            </thead>
                            <tbody>
                            {fichas.map((f) => (
                                <tr key={f.id}>
                                    <td>{f.fullName}</td>
                                    <td>{f.requestType}</td>
                                    <td>{f.acs}</td>
                                    <td>{formatDateTime(f.createdAt)}</td>
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    )}
                </section>
            </main>
        </div>
    )
}

export default App
