exports.up = async (knex) => {
  await knex.schema.createTable('tags', (t) => {
    t.increments('id').primary();
    t.string('name').notNullable().unique();
    t.timestamps(true, true);
  });
  await knex.raw('CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name ON tags (name COLLATE NOCASE)');

  await knex.schema.createTable('video_tags', (t) => {
    t.integer('video_id').notNullable().references('id').inTable('videos').onDelete('CASCADE');
    t.integer('tag_id').notNullable().references('id').inTable('tags').onDelete('CASCADE');
    t.primary(['video_id', 'tag_id']);
  });
  await knex.raw('CREATE INDEX IF NOT EXISTS idx_vt_video ON video_tags (video_id)');
  await knex.raw('CREATE INDEX IF NOT EXISTS idx_vt_tag ON video_tags (tag_id)');
};

exports.down = async (knex) => {
  await knex.schema.dropTable('video_tags');
  await knex.schema.dropTable('tags');
};
