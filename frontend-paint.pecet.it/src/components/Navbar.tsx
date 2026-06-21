import { Link } from "react-router";

export function Navbar() {
    return (
        <nav className="flex gap-4 p-1 px-4 bg-gray-300
        fon border-b-2 font-mono font-extrabold  border-l-2 fixed right-0
         justify-center  rounded-bl-xl">
            <Link to="/">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" id="Home-Simple--Streamline-Majesticons" height="32" width="32">
                    <desc>
                        Home Simple Streamline Icon: https://streamlinehq.com
                    </desc>
                    <path fill="#000000" stroke="#000000" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.3333" d="M13.333333333333332 12.666666666666666v-5.666666666666666a0.6666666666666666 0.6666666666666666 0 0 0 -0.26666666666666666 -0.5333333333333333l-4.666666666666666 -3.5a0.6666666666666666 0.6666666666666666 0 0 0 -0.7999999999999999 0l-4.666666666666666 3.5a0.6666666666666666 0.6666666666666666 0 0 0 -0.26666666666666666 0.5333333333333333V12.666666666666666a0.6666666666666666 0.6666666666666666 0 0 0 0.6666666666666666 0.6666666666666666h9.333333333333332a0.6666666666666666 0.6666666666666666 0 0 0 0.6666666666666666 -0.6666666666666666z"></path>
                </svg>
            </Link>
            <Link to="/photos">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" id="Camera--Streamline-Majesticons" height="32" width="32">
                    <desc>
                        Camera Streamline Icon: https://streamlinehq.com
                    </desc>
                    <path fill="#000000" fill-rule="evenodd" d="M5.049333333333333 2.8906666666666667A2 2 0 0 1 6.713333333333333 2h2.5733333333333333a2 2 0 0 1 1.664 0.8906666666666667l0.5413333333333333 0.8126666666666666A0.6666666666666666 0.6666666666666666 0 0 0 12.046666666666667 4H12.666666666666666a2 2 0 0 1 2 2v6a2 2 0 0 1 -2 2H3.333333333333333a2 2 0 0 1 -2 -2V6a2 2 0 0 1 2 -2h0.62a0.6666666666666666 0.6666666666666666 0 0 0 0.5546666666666666 -0.29666666666666663l0.5413333333333333 -0.8133333333333332zM6.666666666666666 8.666666666666666a1.3333333333333333 1.3333333333333333 0 1 1 2.6666666666666665 0 1.3333333333333333 1.3333333333333333 0 0 1 -2.6666666666666665 0zm1.3333333333333333 -2.6666666666666665a2.6666666666666665 2.6666666666666665 0 1 0 0 5.333333333333333 2.6666666666666665 2.6666666666666665 0 0 0 0 -5.333333333333333z" clip-rule="evenodd" stroke-width="0.6667"></path>
                </svg>
            </Link>
            {/* <Link to="/music">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" id="Music--Streamline-Majesticons" height="32" width="32">
                    <desc>
                        Music Streamline Icon: https://streamlinehq.com
                    </desc>
                    <path fill="#000000" stroke="#000000" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.3333" d="M2 12a2 2 0 1 0 4 0 2 2 0 1 0 -4 0"></path>
                    <path fill="#000000" stroke="#000000" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.3333" d="M10 11.333333333333332a2 2 0 1 0 4 0 2 2 0 1 0 -4 0"></path>
                    <path fill="#000000" d="M14 2 6 4v2.6666666666666665l8 -2V2z" stroke-width="0.6667"></path>
                    <path stroke="#000000" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.3333" d="M6 12v-5.333333333333333m8 4.666666666666666V4.666666666666666M6 6.666666666666666V4l8 -2v2.6666666666666665M6 6.666666666666666l8 -2"></path>
                </svg>
            </Link> */}
        </nav>
    );
}

export default Navbar;   