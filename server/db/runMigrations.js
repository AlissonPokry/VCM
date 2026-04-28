const db = require('./connection');

async function runMigrations() {
  await db.migrate.latest();
  console.log('Database migrations complete');
}

if (require.main === module) {
  runMigrations()
    .then(() => db.destroy())
    .catch(async (error) => {
      console.error(error);
      await db.destroy();
      process.exit(1);
    });
}

module.exports = runMigrations;
