import {
  drizzle as drizzleNode,
  type NodePgDatabase,
} from "drizzle-orm/node-postgres";
import { Pool as NodePool } from "pg";
import { authEnv } from "../authEnv";
import type { relations } from "./relations";

let cachedDB: NodePgDatabase<typeof relations> | null = null;

const getAuthDB = () => {
  if (null === cachedDB) {
    const connectionString = authEnv.AUTH_DB_URL;

    const client = new NodePool({ connectionString });
    cachedDB = drizzleNode({
      client,
    });
  }

  return cachedDB;
};

export const authDB = new Proxy({} as NodePgDatabase<typeof relations>, {
  get(_target, prop, receiver) {
    return Reflect.get(getAuthDB(), prop, receiver);
  },
});
