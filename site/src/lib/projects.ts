import fs from 'node:fs';
import path from 'node:path';

interface Project {
  id: string;
  name: string;
  description: string;
}

interface ProjectData {
  projects: Project[];
}

const projectsPath = path.resolve(process.cwd(), 'data', '_projects.json');

let cached: ProjectData | null = null;

function loadProjectData(): ProjectData {
  if (cached) return cached;
  const raw = fs.readFileSync(projectsPath, 'utf-8');
  cached = JSON.parse(raw) as ProjectData;
  return cached;
}

export function loadProjects(): Project[] {
  return loadProjectData().projects;
}

export function getProjectMeta(projectId: string): Project | undefined {
  return loadProjects().find((p) => p.id === projectId);
}

export function getAllProjectIds(): string[] {
  return loadProjects().map((p) => p.id);
}
