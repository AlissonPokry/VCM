exports.up = (knex) => knex.schema.createTable('videos', (t) => {
  t.increments('id').primary();
  t.string('title').notNullable();
  t.text('description');
  t.string('filename').notNullable().unique();
  t.string('original_name').notNullable();
  t.integer('file_size').notNullable();
  t.integer('duration');
  t.string('thumbnail');
  t.string('platform').defaultTo('instagram');
  t.string('status').defaultTo('scheduled');
  t.datetime('scheduled_at');
  t.datetime('posted_at');
  t.string('n8n_workflow_id');
  t.text('n8n_execution_log').defaultTo('[]');
  t.timestamps(true, true);
});

exports.down = (knex) => knex.schema.dropTable('videos');
