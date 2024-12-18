import {EvaluationContext} from "@openfeature/react-sdk";

export const existingUsers: [{ name: string, ctx: EvaluationContext }] = [
  {
    name: "Anonymous",
    ctx: {targetingKey: "anonymous"}
  },
  {
    name: "User 1",
    ctx: {targetingKey: "user-1", userType: "dev", email: "john.doe@gofeatureflag.org"}
  },
  {
    name: "User 2",
    ctx: {targetingKey: "user-2", userType: "dev", email: "contact@gofeatureflag.org"}
  },
  {
    name: "User 3",
    ctx: {targetingKey: "user-3", userType: "admin", company: "GO Feature Flag"}
  },
  {
    name: "User 4",
    ctx: {targetingKey: "user-4", userType: "customer", location: "Paris"}
  },
  {
    name: "User 5",
    ctx: {targetingKey: "user-5"}
  }
]

export const login = (userName: string): EvaluationContext => {
  return existingUsers.find(user => user.name === userName)?.ctx ?? {targetingKey: "anonymous"}
}