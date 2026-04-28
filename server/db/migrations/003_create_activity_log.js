exports.up = async (knex) => {
  await knex.schema.createTable('activity_log', (t) => {
    t.increments('id').primary();
    t.integer('video_id').references('id').inTable('videos').onDelete('SET NULL');
    t.string('action').notNullable();
    t.text('detail');
    t.string('source').defaultTo('user');
    t.timestamp('created_at').defaultTo(knex.fn.now());
  });
  await knex.raw('CREATE INDEX IF NOT EXISTS idx_act_video ON activity_log (video_id)');
  await knex.raw('CREATE INDEX IF NOT EXISTS idx_act_created ON activity_log (created_at DESC)');
};

exports.down = (knex) => knex.schema.dropTable('activity_log');
