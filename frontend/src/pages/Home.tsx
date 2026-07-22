import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router';
import type { RoomConfig, RoomInfo } from '../gengotypes';



export const Home: React.FC = () => {
  let navigate = useNavigate();

  const [rooms, setRooms] = useState<RoomInfo[]>([]);
  const [isFormVisible, setIsFormVisible] = useState(false);
  const [formData, setFormData] = useState<RoomConfig>({
    name: '',
    password: '',
    is_temporary: true,
    height: 100,
    width: 100,
    is_synth: false,
    is_webrtc: true,
  });

  const fetchRooms = async () => {
    try {
      const response = await fetch('/api/rooms');
      if (response.ok) {
        const data = await response.json();
        setRooms(data);
      }
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    fetchRooms();
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : (type === 'number' || type === 'range') ? Number(value) : value,
    }));
  };

  const handleCreateRoom = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/rooms', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        setIsFormVisible(false);
        setFormData({ name: '', password: '', is_temporary: false, width: 100, height: 100, is_synth: false, is_webrtc: true });
        fetchRooms();
        navigate(`/room/${formData.name}`)
      }
    } catch (error) {
      console.error(error);
    }
  };

  const permanentRooms = rooms.filter((room) => !room.is_temporary);
  const temporaryRooms = rooms.filter((room) => room.is_temporary);

  const RoomCard = ({ room }: { room: RoomInfo }) => (
    <div className="flex items-center justify-between p-4 mb-3 bg-white border border-gray-200 rounded-xl shadow-sm hover:shadow-md transition-shadow duration-200">
      <div className="flex items-center gap-3">
        <h3 className="text-lg font-semibold text-gray-800">{room.name}</h3>
        {room.is_password && (
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
          </svg>
        )}
        <span className="flex items-center gap-1.5 text-sm font-medium text-emerald-600 bg-emerald-50 px-2.5 py-0.5 rounded-full">
          <div className="w-2 h-2 rounded-full bg-emerald-500"></div>
          {room.online_users} online
        </span>
      </div>
      <Link
        to={`/room/${room.name}`}
        className="px-5 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 focus:ring-4 focus:outline-none focus:ring-blue-300 transition-colors"
      >
        Join
      </Link>
    </div>
  );

  return (
    <div className="tile w-full max-w-md">
      <div className="flex justify-between items-center mb-8 gap-4">
        < h1 className="text-3xl font-extrabold text-gray-900 tracking-tight" > Rooms</h1 >
        <button
          onClick={() => setIsFormVisible(!isFormVisible)}
          className={`btn bg-slate-900`}
        >
          {isFormVisible ? 'Back to List' : 'Create'}
        </button>
      </div >

      {
        isFormVisible ? (
          <form onSubmit={handleCreateRoom} className="flex flex-col m-auto w-full gap-2" >
            <div>
              <label className="">Room Name</label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                required
                className="inpt w-full"
                placeholder="Enter room name..."
              />
            </div>

            <div>
              <label className="">Password <span className="">(Optional)</span></label>
              <input
                type="text"
                name="password"
                value={formData.password}
                onChange={handleInputChange}
                className="inpt w-full"
                placeholder="Leave blank for public room"
              />
            </div>

            {/* <label className="">
              <input
                type="checkbox"
                name="is_temporary"
                checked={formData.is_temporary}
                onChange={handleInputChange}
                className=""
              />
              <span className="text-sm font-semibold ">Temporary Room</span>
            </label> */}
            <label className="flex items-center">
              Width {formData.width}px
              <input type="range" name="width" min="100" max="2000" value={formData.width} onChange={handleInputChange}
                className="" />
            </label>
            <label className="flex items-center">
              Height {formData.height}px
              <input type="range" min="100" name="height" max="2000" value={formData.height} onChange={handleInputChange}
                className="" />
            </label>
            <button
              type="submit"
              className="btn bg-black"
            >
              Create
            </button>
          </form>
        ) : (
          <div className="flex flex-col gap-8">
            <div>
              {permanentRooms.length > 0 ? (
                permanentRooms.map((room) => <RoomCard key={room.name} room={room} />)
              ) : (
                <p>No rooms available.</p>
              )}
            </div>

            <div>
              {temporaryRooms.length > 0 ? (
                temporaryRooms.map((room) => <RoomCard key={room.name} room={room} />)
              ) : (
                permanentRooms.length == 0 && <p>No rooms available.</p>
              )}
            </div>
          </div>
        )}
    </div >
  );
};