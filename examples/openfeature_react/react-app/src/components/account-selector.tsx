import {existingUsers} from "../service/login.tsx";

export const AccountSelector = ({onChange}: { onChange: (name: string) => void }) => {
  return <div>
    <form className="max-w-md mx-auto flex">
      <span className={"min-w-20 pt-2 text-gray-200 "}>Login as:</span>
      <select id="users"
              onChange={() => onChange((document.getElementById("users") as HTMLSelectElement).value)}
              className="min-w-xs bg-gray-50 border border-gray-300 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block w-full p-2.5 dark:bg-zinc-700 dark:border-zinc-600 dark:placeholder-zinc-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500">
        {
          existingUsers.map((user) => {
            return <option key={user.name} value={user.name}>{user.name}</option>
          })
        }
      </select>
    </form>
  </div>
}