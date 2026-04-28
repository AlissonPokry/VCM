const app = require('./app');
const runMigrations = require('./db/runMigrations');

const port = Number(process.env.PORT || 3001);

async function start() {
  await runMigrations();
  app.listen(port, () => {
    console.log(`V.C.M API listening on http://localhost:${port}`);
  });
}

start().catch((error) => {
  console.error(error);
  process.exit(1);
});
