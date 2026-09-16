import {useEffect, useState} from 'react';
import './App.css';
import {
    DeleteACS,
    DeleteFicha,
    ListACS,
    ListAvailableMonths,
    ListFichas,
    ListFichasByACS,
    ListRequestTypes,
    RegisterACS,
    RegisterFicha,
    StatsByACS,
    StatsByRequestType,
    UpdateACS,
    UpdateFicha,
} from "../wailsjs/go/main/App";
import {BrowserOpenURL} from "../wailsjs/runtime";
import Login from './Login';

const emptyFichaForm = {fullName: '', requestType: '', requestTypeOther: '', acsId: '', phone: '', notified: true};
const emptyACSForm = {name: '', phone: ''};

function currentMonth() {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
}

function formatDateTime(isoString) {
    const date = new Date(isoString);
    if (isNaN(date.getTime())) return isoString;
    return date.toLocaleString('pt-BR');
}

function formatMonthLabel(month) {
    const [year, monthNumber] = month.split('-');
    const date = new Date(Number(year), Number(monthNumber) - 1, 1);
    const label = date.toLocaleDateString('pt-BR', {month: 'long', year: 'numeric'});
    return label.charAt(0).toUpperCase() + label.slice(1);
}

function whatsAppURL(phone, message) {
    const digits = phone.replace(/\D/g, '');
    const withCountryCode = digits.startsWith('55') ? digits : `55${digits}`;
    return `https://wa.me/${withCountryCode}?text=${encodeURIComponent(message)}`;
}

function defaultWhatsAppMessage(ficha) {
    return `Olá ${ficha.fullName}, informamos que sua solicitação de ${ficha.requestType} já está disponível. Por favor, entre em contato conosco.`;
}

function BarStats({items, emptyLabel}) {
    if (!items || items.length === 0) {
        return <p className="empty">{emptyLabel}</p>;
    }
    const max = Math.max(...items.map((item) => item.count), 1);
    return (
        <div className="bar-stats">
            {items.map((item) => (
                <div className="stat-bar-row" key={item.label}>
                    <span className="stat-bar-label" title={item.label}>{item.label}</span>
                    <span className="stat-bar-track">
                        <span className="stat-bar-fill" style={{width: `${(item.count / max) * 100}%`}}/>
                    </span>
                    <span className="stat-bar-count">{item.count}</span>
                </div>
            ))}
        </div>
    );
}

function App() {
    const [loggedIn, setLoggedIn] = useState(false);

    if (!loggedIn) {
        return <Login onAuthenticated={() => setLoggedIn(true)}/>;
    }

    return <MainApp/>;
}

function MainApp() {
    const [activeTab, setActiveTab] = useState('fichas');

    const [requestTypes, setRequestTypes] = useState([]);
    const [acsList, setAcsList] = useState([]);

    const [fichaForm, setFichaForm] = useState(emptyFichaForm);
    const [editingFichaId, setEditingFichaId] = useState(null);
    const [fichas, setFichas] = useState([]);
    const [query, setQuery] = useState('');
    const [months, setMonths] = useState([currentMonth()]);
    const [selectedMonth, setSelectedMonth] = useState(currentMonth());
    const [fichaError, setFichaError] = useState('');
    const [fichaSuccess, setFichaSuccess] = useState('');
    const [fichaSubmitting, setFichaSubmitting] = useState(false);
    const [duplicateNotice, setDuplicateNotice] = useState('');

    const [acsForm, setAcsForm] = useState(emptyACSForm);
    const [editingACSId, setEditingACSId] = useState(null);
    const [acsError, setAcsError] = useState('');
    const [acsSuccess, setAcsSuccess] = useState('');
    const [acsSubmitting, setAcsSubmitting] = useState(false);

    const [examsACS, setExamsACS] = useState(null);
    const [examsFichas, setExamsFichas] = useState([]);

    const [statsStart, setStatsStart] = useState('');
    const [statsEnd, setStatsEnd] = useState('');
    const [requestTypeStats, setRequestTypeStats] = useState([]);
    const [acsStats, setAcsStats] = useState([]);
    const [statsError, setStatsError] = useState('');

    const loadRequestTypes = () => {
        ListRequestTypes().then(setRequestTypes).catch((err) => setFichaError(String(err)));
    };

    const loadACSList = () => {
        ListACS().then(setAcsList).catch((err) => setFichaError(String(err)));
    };

    const loadMonths = () => {
        ListAvailableMonths()
            .then((found) => {
                const merged = Array.from(new Set([currentMonth(), ...(found || [])]));
                merged.sort().reverse();
                setMonths(merged);
            })
            .catch((err) => setFichaError(String(err)));
    };

    const loadFichas = (name, month) => {
        ListFichas(name, month).then(setFichas).catch((err) => setFichaError(String(err)));
    };

    useEffect(() => {
        loadRequestTypes();
        loadACSList();
        loadMonths();
        loadFichas('', selectedMonth);
    }, []);

    useEffect(() => {
        const timeout = setTimeout(() => loadFichas(query.trim(), selectedMonth), 250);
        return () => clearTimeout(timeout);
    }, [query, selectedMonth]);

    useEffect(() => {
        const name = fichaForm.fullName.trim();
        if (!name) {
            setDuplicateNotice('');
            return;
        }
        const timeout = setTimeout(() => {
            ListFichas(name, '')
                .then((found) => {
                    const match = (found || []).some(
                        (f) => f.fullName.toLowerCase() === name.toLowerCase() && f.id !== editingFichaId
                    );
                    setDuplicateNotice(match ? `Já existe uma ficha registrada para "${name}".` : '');
                })
                .catch(() => setDuplicateNotice(''));
        }, 300);
        return () => clearTimeout(timeout);
    }, [fichaForm.fullName, editingFichaId]);

    const updateFichaField = (field) => (e) => {
        const value = e.target.type === 'checkbox' ? e.target.checked : e.target.value;
        setFichaForm((prev) => ({...prev, [field]: value}));
    };

    const resetFichaForm = () => {
        setFichaForm(emptyFichaForm);
        setEditingFichaId(null);
    };

    const startEditFicha = (f) => {
        const isKnownType = requestTypes.includes(f.requestType) && f.requestType !== 'Outro';
        setFichaForm({
            fullName: f.fullName,
            requestType: isKnownType ? f.requestType : 'Outro',
            requestTypeOther: isKnownType ? '' : f.requestType,
            acsId: f.acsId,
            phone: f.phone,
            notified: f.notified,
        });
        setEditingFichaId(f.id);
        setFichaError('');
        setFichaSuccess('');
    };

    const deleteFicha = (f) => {
        if (!window.confirm(`Apagar a ficha de ${f.fullName}?`)) return;
        DeleteFicha(f.id)
            .then(() => {
                if (editingFichaId === f.id) resetFichaForm();
                loadMonths();
                loadFichas(query.trim(), selectedMonth);
            })
            .catch((err) => setFichaError(String(err)));
    };

    const sendWhatsApp = (f) => {
        BrowserOpenURL(whatsAppURL(f.phone, defaultWhatsAppMessage(f)));
    };

    const submitFicha = (e) => {
        e.preventDefault();
        setFichaError('');
        setFichaSuccess('');

        const requestType = fichaForm.requestType === 'Outro'
            ? fichaForm.requestTypeOther.trim()
            : fichaForm.requestType;

        if (!fichaForm.acsId) {
            setFichaError('Selecione um ACS.');
            return;
        }
        if (!requestType) {
            setFichaError('Informe o tipo de solicitação.');
            return;
        }

        setFichaSubmitting(true);
        const call = editingFichaId
            ? UpdateFicha(editingFichaId, fichaForm.fullName.trim(), requestType, fichaForm.acsId, fichaForm.phone.trim(), fichaForm.notified)
            : RegisterFicha(fichaForm.fullName.trim(), requestType, fichaForm.acsId, fichaForm.phone.trim(), fichaForm.notified);

        call
            .then((ficha) => {
                setFichaSuccess(editingFichaId
                    ? `Ficha de ${ficha.fullName} atualizada.`
                    : `Ficha de ${ficha.fullName} registrada com sucesso.`);
                resetFichaForm();
                loadMonths();
                loadFichas(query.trim(), selectedMonth);
                if (examsACS) openACSExams(examsACS);
            })
            .catch((err) => setFichaError(String(err)))
            .finally(() => setFichaSubmitting(false));
    };

    const updateACSField = (field) => (e) => {
        setAcsForm((prev) => ({...prev, [field]: e.target.value}));
    };

    const resetACSForm = () => {
        setAcsForm(emptyACSForm);
        setEditingACSId(null);
    };

    const startEditACS = (acs) => {
        setAcsForm({name: acs.name, phone: acs.phone});
        setEditingACSId(acs.id);
        setAcsError('');
        setAcsSuccess('');
    };

    const deleteACS = (acs) => {
        if (!window.confirm(`Apagar o ACS ${acs.name}?`)) return;
        DeleteACS(acs.id)
            .then(() => {
                if (editingACSId === acs.id) resetACSForm();
                if (examsACS && examsACS.id === acs.id) setExamsACS(null);
                loadACSList();
            })
            .catch((err) => setAcsError(String(err)));
    };

    const submitACS = (e) => {
        e.preventDefault();
        setAcsError('');
        setAcsSuccess('');
        setAcsSubmitting(true);

        const call = editingACSId
            ? UpdateACS(editingACSId, acsForm.name.trim(), acsForm.phone.trim())
            : RegisterACS(acsForm.name.trim(), acsForm.phone.trim());

        call
            .then((acs) => {
                setAcsSuccess(editingACSId ? `ACS ${acs.name} atualizado.` : `ACS ${acs.name} cadastrado.`);
                resetACSForm();
                loadACSList();
            })
            .catch((err) => setAcsError(String(err)))
            .finally(() => setAcsSubmitting(false));
    };

    const openACSExams = (acs) => {
        setExamsACS(acs);
        ListFichasByACS(acs.id).then(setExamsFichas).catch((err) => setAcsError(String(err)));
    };

    useEffect(() => {
        if (activeTab !== 'analises') return;
        setStatsError('');
        StatsByRequestType(statsStart, statsEnd).then(setRequestTypeStats).catch((err) => setStatsError(String(err)));
        StatsByACS(statsStart, statsEnd).then(setAcsStats).catch((err) => setStatsError(String(err)));
    }, [activeTab, statsStart, statsEnd]);

    return (
        <div id="App">
            <header className="header">
                <h1>Ficha Tracker</h1>
                <p>Impressão de solicitações</p>
            </header>

            <nav className="tabs">
                <button
                    className={`tab ${activeTab === 'fichas' ? 'active' : ''}`}
                    onClick={() => setActiveTab('fichas')}
                >
                    Fichas
                </button>
                <button
                    className={`tab ${activeTab === 'acs' ? 'active' : ''}`}
                    onClick={() => setActiveTab('acs')}
                >
                    ACS
                </button>
                <button
                    className={`tab ${activeTab === 'analises' ? 'active' : ''}`}
                    onClick={() => setActiveTab('analises')}
                >
                    Análises
                </button>
            </nav>

            {activeTab === 'fichas' && (
                <main className="content">
                    <section className="card">
                        <h2>{editingFichaId ? 'Editar ficha' : 'Registrar ficha'}</h2>
                        <form onSubmit={submitFicha} className="form">
                            <label>
                                Nome completo
                                <input
                                    type="text"
                                    value={fichaForm.fullName}
                                    onChange={updateFichaField('fullName')}
                                    autoComplete="off"
                                    required
                                />
                            </label>
                            {duplicateNotice && <p className="message warning">{duplicateNotice}</p>}
                            <label>
                                Telefone do paciente
                                <input
                                    type="tel"
                                    value={fichaForm.phone}
                                    onChange={updateFichaField('phone')}
                                    placeholder="(11) 99999-9999"
                                    autoComplete="off"
                                />
                            </label>
                            <label>
                                Tipo de solicitação
                                <select value={fichaForm.requestType} onChange={updateFichaField('requestType')} required>
                                    <option value="" disabled>Selecione...</option>
                                    {requestTypes.map((type) => (
                                        <option key={type} value={type}>{type}</option>
                                    ))}
                                </select>
                            </label>
                            {fichaForm.requestType === 'Outro' && (
                                <label>
                                    Qual?
                                    <input
                                        type="text"
                                        value={fichaForm.requestTypeOther}
                                        onChange={updateFichaField('requestTypeOther')}
                                        required
                                    />
                                </label>
                            )}
                            <label>
                                ACS
                                <select value={fichaForm.acsId} onChange={updateFichaField('acsId')} required>
                                    <option value="" disabled>Selecione...</option>
                                    {acsList.map((acs) => (
                                        <option key={acs.id} value={acs.id}>{acs.name}</option>
                                    ))}
                                </select>
                            </label>
                            <label>
                                Paciente avisado?
                                <select
                                    value={fichaForm.notified ? 'sim' : 'nao'}
                                    onChange={(e) => setFichaForm((prev) => ({...prev, notified: e.target.value === 'sim'}))}
                                >
                                    <option value="sim">Sim</option>
                                    <option value="nao">Não</option>
                                </select>
                            </label>
                            <div className="form-actions">
                                <button type="submit" className="btn" disabled={fichaSubmitting}>
                                    {fichaSubmitting ? 'Salvando...' : (editingFichaId ? 'Salvar alterações' : 'Registrar')}
                                </button>
                                {editingFichaId && (
                                    <button type="button" className="btn btn-secondary" onClick={resetFichaForm}>
                                        Cancelar
                                    </button>
                                )}
                            </div>
                        </form>
                        {fichaError && <p className="message error">{fichaError}</p>}
                        {fichaSuccess && <p className="message success">{fichaSuccess}</p>}
                    </section>

                    <section className="card">
                        <div className="list-header">
                            <h2>Fichas registradas</h2>
                            <div className="list-filters">
                                <select
                                    className="month-select"
                                    value={selectedMonth}
                                    onChange={(e) => setSelectedMonth(e.target.value)}
                                >
                                    {months.map((month) => (
                                        <option key={month} value={month}>
                                            {formatMonthLabel(month)}
                                        </option>
                                    ))}
                                </select>
                                <input
                                    type="text"
                                    className="search"
                                    placeholder="Buscar por nome..."
                                    value={query}
                                    onChange={(e) => setQuery(e.target.value)}
                                />
                            </div>
                        </div>

                        {fichas.length === 0 ? (
                            <p className="empty">Nenhuma ficha encontrada.</p>
                        ) : (
                            <div className="table-wrap">
                                <table className="table">
                                    <thead>
                                    <tr>
                                        <th>Nome</th>
                                        <th>Tipo de solicitação</th>
                                        <th>ACS</th>
                                        <th>Telefone</th>
                                        <th>Avisado</th>
                                        <th>Registrado em</th>
                                        <th>Ações</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                    {fichas.map((f) => (
                                        <tr key={f.id}>
                                            <td>{f.fullName}</td>
                                            <td>{f.requestType}</td>
                                            <td>{f.acsName}</td>
                                            <td>{f.phone}</td>
                                            <td>{f.notified ? 'Sim' : 'Não'}</td>
                                            <td>{formatDateTime(f.createdAt)}</td>
                                            <td className="actions">
                                                {!f.notified && f.phone && (
                                                    <button
                                                        type="button"
                                                        className="icon-btn whatsapp"
                                                        title="Avisar pelo WhatsApp"
                                                        onClick={() => sendWhatsApp(f)}
                                                    >
                                                        WhatsApp
                                                    </button>
                                                )}
                                                <button type="button" className="icon-btn" onClick={() => startEditFicha(f)}>
                                                    Editar
                                                </button>
                                                <button type="button" className="icon-btn danger" onClick={() => deleteFicha(f)}>
                                                    Apagar
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </section>
                </main>
            )}

            {activeTab === 'acs' && (
                <main className="content">
                    <section className="card">
                        <h2>{editingACSId ? 'Editar ACS' : 'Cadastrar ACS'}</h2>
                        <form onSubmit={submitACS} className="form">
                            <label>
                                Nome
                                <input
                                    type="text"
                                    value={acsForm.name}
                                    onChange={updateACSField('name')}
                                    autoComplete="off"
                                    required
                                />
                            </label>
                            <label>
                                Telefone
                                <input
                                    type="tel"
                                    value={acsForm.phone}
                                    onChange={updateACSField('phone')}
                                    placeholder="(11) 99999-9999"
                                    autoComplete="off"
                                    required
                                />
                            </label>
                            <div className="form-actions">
                                <button type="submit" className="btn" disabled={acsSubmitting}>
                                    {acsSubmitting ? 'Salvando...' : (editingACSId ? 'Salvar alterações' : 'Cadastrar')}
                                </button>
                                {editingACSId && (
                                    <button type="button" className="btn btn-secondary" onClick={resetACSForm}>
                                        Cancelar
                                    </button>
                                )}
                            </div>
                        </form>
                        {acsError && <p className="message error">{acsError}</p>}
                        {acsSuccess && <p className="message success">{acsSuccess}</p>}
                    </section>

                    <section className="card">
                        <h2>ACS cadastrados</h2>
                        {acsList.length === 0 ? (
                            <p className="empty">Nenhum ACS cadastrado.</p>
                        ) : (
                            <div className="table-wrap">
                                <table className="table">
                                    <thead>
                                    <tr>
                                        <th>Nome</th>
                                        <th>Telefone</th>
                                        <th>Cadastrado em</th>
                                        <th>Ações</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                    {acsList.map((acs) => (
                                        <tr key={acs.id}>
                                            <td>
                                                <button type="button" className="link-btn" onClick={() => openACSExams(acs)}>
                                                    {acs.name}
                                                </button>
                                            </td>
                                            <td>{acs.phone}</td>
                                            <td>{formatDateTime(acs.createdAt)}</td>
                                            <td className="actions">
                                                <button type="button" className="icon-btn" onClick={() => startEditACS(acs)}>
                                                    Editar
                                                </button>
                                                <button type="button" className="icon-btn danger" onClick={() => deleteACS(acs)}>
                                                    Apagar
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </section>

                    {examsACS && (
                        <section className="card printable">
                            <div className="list-header no-print">
                                <h2>Exames de {examsACS.name}</h2>
                                <div className="list-filters">
                                    <button type="button" className="btn" onClick={() => window.print()}>
                                        Imprimir
                                    </button>
                                    <button type="button" className="btn btn-secondary" onClick={() => setExamsACS(null)}>
                                        Fechar
                                    </button>
                                </div>
                            </div>
                            <h2 className="print-only">Exames de {examsACS.name}</h2>
                            {examsFichas.length === 0 ? (
                                <p className="empty">Nenhum exame registrado para este ACS.</p>
                            ) : (
                                <div className="table-wrap">
                                    <table className="table">
                                        <thead>
                                        <tr>
                                            <th>Nome do paciente</th>
                                            <th>Tipo de solicitação</th>
                                            <th>Registrado em</th>
                                        </tr>
                                        </thead>
                                        <tbody>
                                        {examsFichas.map((f) => (
                                            <tr key={f.id}>
                                                <td>{f.fullName}</td>
                                                <td>{f.requestType}</td>
                                                <td>{formatDateTime(f.createdAt)}</td>
                                            </tr>
                                        ))}
                                        </tbody>
                                    </table>
                                </div>
                            )}
                        </section>
                    )}
                </main>
            )}

            {activeTab === 'analises' && (
                <main className="content">
                    <section className="card">
                        <h2>Janela de análise</h2>
                        <div className="date-filters">
                            <label>
                                De
                                <input
                                    type="date"
                                    value={statsStart}
                                    onChange={(e) => setStatsStart(e.target.value)}
                                />
                            </label>
                            <label>
                                Até
                                <input
                                    type="date"
                                    value={statsEnd}
                                    onChange={(e) => setStatsEnd(e.target.value)}
                                />
                            </label>
                            {(statsStart || statsEnd) && (
                                <button
                                    type="button"
                                    className="btn btn-secondary"
                                    onClick={() => {
                                        setStatsStart('');
                                        setStatsEnd('');
                                    }}
                                >
                                    Todo o período
                                </button>
                            )}
                        </div>
                        {statsError && <p className="message error">{statsError}</p>}
                    </section>

                    <section className="card">
                        <h2>Exames por tipo de solicitação</h2>
                        <BarStats
                            items={requestTypeStats.map((s) => ({label: s.requestType, count: s.count}))}
                            emptyLabel="Nenhum exame no período selecionado."
                        />
                    </section>

                    <section className="card">
                        <h2>Ranking de ACS</h2>
                        <BarStats
                            items={acsStats.map((s) => ({label: s.acsName || '(sem ACS)', count: s.count}))}
                            emptyLabel="Nenhum exame no período selecionado."
                        />
                    </section>
                </main>
            )}
        </div>
    )
}

export default App
