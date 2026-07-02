import React, { useState } from 'react';
import { useNavigate } from 'react-router';
import { useStore } from '../Store';

export const LoginForm: React.FC<{ isPassword?: boolean }> = ({ isPassword }) => {
    const [name, setName] = useState<string>('');
    const [password, setPassword] = useState<string>('');
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [message, setMessage] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);

    const navigate = useNavigate();
    const { checkAuth } = useStore();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!name.trim()) {
            setError('Name cannot be empty');
            return;
        }

        setIsLoading(true);
        setError(null);
        setMessage(null);

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name: name.trim(), password: password.trim() }),
                credentials: 'include',
            });

            const data = await response.text();

            if (!response.ok) {
                throw new Error(data || `Server responded with status ${response.status}`);
            }

            setMessage(data); // "Logged successful"
            setName('');
            setPassword('');
            await checkAuth();

            navigate('/');
        } catch (err) {
            if (err instanceof Error) {
                setError(err.message);
            } else {
                setError('An unexpected error occurred');
            }
        } finally {
            setIsLoading(false);
        }
    };
    return (
        <div className="max-w-[320px] font-sans">

            <form onSubmit={handleSubmit} className="flex flex-col gap-3">
                <div className="flex flex-col gap-2">
                    <input
                        id="name"
                        type="text"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        disabled={isLoading}
                        className="w-full p-2 border border-gray-300 rounded box-border
                         focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100 disabled:text-gray-400"
                        placeholder="Name"
                    />
                    {isPassword && (
                        <input
                            id="password"
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            disabled={isLoading}
                            className="w-full p-2 border border-gray-300 rounded
                             box-border focus:outline-none focus:ring-2 focus:ring-blue-500
                              disabled:bg-gray-100 disabled:text-gray-400"
                            placeholder="Password"
                        />
                    )}
                </div>

                <button
                    type="submit"
                    disabled={isLoading}
                    className="p-2.5 bg-blue-600 text-white font-medium rounded transition hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed cursor-pointer"
                >
                    {isLoading ? 'Logging in...' : 'Submit'}
                </button>
            </form>

            {error && (
                <div className="text-red-600 text-sm mt-3">
                    <strong>Error:</strong> {error}
                </div>
            )}

            {message && (
                <div className="text-green-600 text-sm mt-3">
                    {message}
                </div>
            )}
        </div>
    );
};

export const Login = () => {
    const [isPassword, setIsPassword] = useState<boolean>(false);
    return (
        <>
            <div className='bg-gray-200 p-4 rounded-lg m-auto'>
                {isPassword ? <LoginForm isPassword={true} /> : <LoginForm />}

            </div>
            <button onClick={() => setIsPassword(!isPassword)} className="">
                {isPassword ? 'Normal Login' : 'Admin Login'}
            </button>
        </>
    )
}
export const LoginAdmin = () => {
    return (
        <div className='bg-gray-200 p-4 rounded-lg m-auto'>
            <LoginForm isPassword={true} />
        </div>
    )
}