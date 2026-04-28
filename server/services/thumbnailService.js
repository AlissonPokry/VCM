const fs = require('fs');
const path = require('path');
const { execFile } = require('child_process');
const ffmpeg = require('fluent-ffmpeg');
const ffmpegStatic = require('ffmpeg-static');

ffmpeg.setFfmpegPath(ffmpegStatic);

const resolveProjectPath = (value, fallback) => {
  const target = value || fallback;
  return path.isAbsolute(target) ? target : path.resolve(__dirname, '../../', target);
};

const thumbnailDir = resolveProjectPath(process.env.THUMBNAIL_DIR, './server/thumbnails');
fs.mkdirSync(thumbnailDir, { recursive: true });

function probeDuration(inputPath) {
  return new Promise((resolve) => {
    execFile(ffmpegStatic, ['-i', inputPath], (error, stdout, stderr) => {
      const output = `${stdout}\n${stderr}`;
      const match = output.match(/Duration:\s(\d{2}):(\d{2}):(\d{2}\.\d{2})/);
      if (!match) {
        if (error) console.error(error.message);
        return resolve(null);
      }
      const [, hours, minutes, seconds] = match;
      const total = Number(hours) * 3600 + Number(minutes) * 60 + Number(seconds);
      return resolve(Math.round(total));
    });
  });
}

function extractFrame(inputPath, outputPath, duration) {
  return new Promise((resolve) => {
    const seek = duration && duration <= 1 ? '0' : '1';
    execFile(
      ffmpegStatic,
      [
        '-y',
        '-ss',
        seek,
        '-i',
        inputPath,
        '-frames:v',
        '1',
        '-vf',
        'scale=640:360:force_original_aspect_ratio=decrease,pad=640:360:(ow-iw)/2:(oh-ih)/2',
        '-q:v',
        '2',
        outputPath
      ],
      (error) => {
        if (error) {
          console.error(error.message);
          return resolve(false);
        }
        return resolve(fs.existsSync(outputPath));
      }
    );
  });
}

async function processVideo(inputPath, outputBasename) {
  try {
    const duration = await probeDuration(inputPath);
    const absoluteOutput = path.join(thumbnailDir, `${outputBasename}.jpg`);
    const extracted = await extractFrame(inputPath, absoluteOutput, duration);
    const thumbnailPath = extracted ? path.relative(path.resolve(__dirname, '../../'), absoluteOutput).replace(/\\/g, '/') : null;
    return { thumbnailPath, duration };
  } catch (error) {
    console.error(error);
    return { thumbnailPath: null, duration: null };
  }
}

module.exports = { processVideo, thumbnailDir };
