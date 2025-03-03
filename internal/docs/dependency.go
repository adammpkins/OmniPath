package docs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// DependencyDocs holds information about a dependency and its documentation URL.
type DependencyDocs struct {
	Name   string
	DocURL string
}

// Helper function to check if a file contains Spring annotations
func hasSpringAnnotations(filePath string) bool {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false
	}

	fileContent := string(content)
	return strings.Contains(fileContent, "@Controller") ||
		strings.Contains(fileContent, "@Service") ||
		strings.Contains(fileContent, "@Repository") ||
		strings.Contains(fileContent, "@Component") ||
		strings.Contains(fileContent, "@SpringBootApplication") ||
		strings.Contains(fileContent, "springframework")
}

// Helper function to check if a C# file is ASP.NET related
func isAspNetFile(filePath string) bool {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false
	}

	fileContent := string(content)
	return strings.Contains(fileContent, "Microsoft.AspNetCore") ||
		strings.Contains(fileContent, "System.Web") ||
		strings.Contains(fileContent, "[ApiController]") ||
		strings.Contains(fileContent, "Controller") ||
		strings.Contains(fileContent, "IActionResult")
}

// DetectDependencies reads various project files
// and returns a list of dependencies along with known documentation URLs.
func DetectDependencies() ([]DependencyDocs, error) {
	// Use a map to prevent duplicate entries
	depsMap := make(map[string]DependencyDocs)

	// Record all file extensions found in the project
	fileExtensions := make(map[string]bool)

	// Walk the entire project directory to gather information
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/directories we can't access
		}

		if !info.IsDir() {
			// Record file extension
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext != "" {
				fileExtensions[ext] = true
			}

			// Check specific files by name or extension
			filename := strings.ToLower(info.Name())

			// Ruby detection
			if ext == ".rb" || ext == ".gemspec" || filename == "gemfile" {
				depsMap["Ruby"] = DependencyDocs{
					Name:   "Ruby",
					DocURL: "https://ruby-doc.org/",
				}
			}

			// Rails detection
			if filename == "gemfile" {
				content, err := ioutil.ReadFile(path)
				if err == nil && strings.Contains(string(content), "rails") {
					depsMap["Ruby on Rails"] = DependencyDocs{
						Name:   "Ruby on Rails",
						DocURL: "https://guides.rubyonrails.org/",
					}
				}
			}

			// Java detection
			if ext == ".java" || ext == ".class" || ext == ".jar" {
				depsMap["Java"] = DependencyDocs{
					Name:   "Java",
					DocURL: "https://docs.oracle.com/en/java/",
				}
			}

			// Spring detection
			if filename == "applicationcontext.xml" || filename == "springconfig.java" ||
				(ext == ".java" && hasSpringAnnotations(path)) {
				depsMap["Spring"] = DependencyDocs{
					Name:   "Spring",
					DocURL: "https://spring.io/projects/spring-framework",
				}
			}

			// Maven/Gradle detection
			if filename == "pom.xml" {
				depsMap["Maven"] = DependencyDocs{
					Name:   "Maven",
					DocURL: "https://maven.apache.org/guides/",
				}
			}
			if filename == "build.gradle" || filename == "build.gradle.kts" {
				depsMap["Gradle"] = DependencyDocs{
					Name:   "Gradle",
					DocURL: "https://docs.gradle.org/",
				}
			}

			// C# detection
			if ext == ".cs" || ext == ".csproj" || ext == ".sln" {
				depsMap["C#"] = DependencyDocs{
					Name:   "C#",
					DocURL: "https://docs.microsoft.com/en-us/dotnet/csharp/",
				}
			}

			// ASP.NET detection
			if ext == ".cshtml" || ext == ".aspx" ||
				(ext == ".cs" && isAspNetFile(path)) {
				depsMap["ASP.NET"] = DependencyDocs{
					Name:   "ASP.NET",
					DocURL: "https://docs.microsoft.com/en-us/aspnet/",
				}
			}

			// TypeScript detection
			if ext == ".ts" || ext == ".tsx" {
				depsMap["TypeScript"] = DependencyDocs{
					Name:   "TypeScript",
					DocURL: "https://www.typescriptlang.org/docs/",
				}
			}

			// Docker detection
			if filename == "dockerfile" || strings.HasPrefix(filename, "docker-compose") {
				depsMap["Docker"] = DependencyDocs{
					Name:   "Docker",
					DocURL: "https://docs.docker.com/",
				}
			}

			// CSS frameworks detection from HTML files
			if ext == ".html" || ext == ".htm" {
				content, err := ioutil.ReadFile(path)
				if err == nil {
					htmlContent := string(content)

					// Bootstrap CDN detection
					if strings.Contains(htmlContent, "bootstrap.min.css") ||
						strings.Contains(htmlContent, "bootstrap.css") ||
						strings.Contains(htmlContent, "maxcdn.bootstrapcdn.com/bootstrap") ||
						strings.Contains(htmlContent, "cdn.jsdelivr.net/npm/bootstrap") ||
						strings.Contains(htmlContent, "stackpath.bootstrapcdn.com/bootstrap") {
						depsMap["Bootstrap"] = DependencyDocs{
							Name:   "Bootstrap",
							DocURL: "https://getbootstrap.com/docs/",
						}
					}

					// jQuery detection
					if strings.Contains(htmlContent, "jquery.min.js") ||
						strings.Contains(htmlContent, "jquery.js") ||
						strings.Contains(htmlContent, "code.jquery.com") {
						depsMap["jQuery"] = DependencyDocs{
							Name:   "jQuery",
							DocURL: "https://api.jquery.com/",
						}
					}

					// Font Awesome detection
					if strings.Contains(htmlContent, "font-awesome.css") ||
						strings.Contains(htmlContent, "fontawesome") ||
						strings.Contains(htmlContent, "fa-") {
						depsMap["Font Awesome"] = DependencyDocs{
							Name:   "Font Awesome",
							DocURL: "https://fontawesome.com/docs",
						}
					}

					// React CDN detection
					if strings.Contains(htmlContent, "react.development.js") ||
						strings.Contains(htmlContent, "react.production.min.js") ||
						strings.Contains(htmlContent, "react-dom") {
						depsMap["React"] = DependencyDocs{
							Name:   "React",
							DocURL: "https://react.dev/reference/react",
						}
					}

					// Vue CDN detection
					if strings.Contains(htmlContent, "vue.js") ||
						strings.Contains(htmlContent, "vue.min.js") {
						depsMap["Vue"] = DependencyDocs{
							Name:   "Vue",
							DocURL: "https://vuejs.org/guide/introduction.html",
						}
					}
				}
			}

			// JavaScript framework detection from JS files
			if ext == ".js" {
				content, err := ioutil.ReadFile(path)
				if err == nil {
					jsContent := string(content)

					// React detection in JS files
					if strings.Contains(jsContent, "React.") ||
						strings.Contains(jsContent, "ReactDOM") ||
						strings.Contains(jsContent, "import React") {
						depsMap["React"] = DependencyDocs{
							Name:   "React",
							DocURL: "https://react.dev/reference/react",
						}
					}

					// Vue detection in JS files
					if strings.Contains(jsContent, "new Vue") ||
						strings.Contains(jsContent, "Vue.component") {
						depsMap["Vue"] = DependencyDocs{
							Name:   "Vue",
							DocURL: "https://vuejs.org/guide/introduction.html",
						}
					}

					// jQuery detection in JS files
					if strings.Contains(jsContent, "$(") ||
						strings.Contains(jsContent, "jQuery") {
						depsMap["jQuery"] = DependencyDocs{
							Name:   "jQuery",
							DocURL: "https://api.jquery.com/",
						}
					}
				}
			}

			// Bootstrap CSS file detection
			if strings.Contains(strings.ToLower(info.Name()), "bootstrap") && strings.HasSuffix(strings.ToLower(info.Name()), ".css") {
				depsMap["Bootstrap"] = DependencyDocs{
					Name:   "Bootstrap",
					DocURL: "https://getbootstrap.com/docs/",
				}
			}
		}
		return nil
	})

	// Add basic language detections based on file extensions
	if fileExtensions[".py"] {
		depsMap["Python"] = DependencyDocs{
			Name:   "Python",
			DocURL: "https://docs.python.org/3/",
		}
	}

	if fileExtensions[".js"] {
		depsMap["JavaScript"] = DependencyDocs{
			Name:   "JavaScript",
			DocURL: "https://developer.mozilla.org/en-US/docs/Web/JavaScript",
		}
	}

	if fileExtensions[".html"] || fileExtensions[".htm"] {
		depsMap["HTML"] = DependencyDocs{
			Name:   "HTML",
			DocURL: "https://developer.mozilla.org/en-US/docs/Web/HTML",
		}
	}

	if fileExtensions[".css"] {
		depsMap["CSS"] = DependencyDocs{
			Name:   "CSS",
			DocURL: "https://developer.mozilla.org/en-US/docs/Web/CSS",
		}
	}

	if fileExtensions[".php"] {
		depsMap["PHP"] = DependencyDocs{
			Name:   "PHP",
			DocURL: "https://www.php.net/docs.php",
		}
	}

	if fileExtensions[".sql"] {
		depsMap["SQL"] = DependencyDocs{
			Name:   "SQL",
			DocURL: "https://www.w3schools.com/sql/",
		}
	}

	// Check for specific configuration files
	configFiles := map[string]struct {
		path string
		name string
		url  string
	}{
		"composer.json": {
			name: "Composer",
			url:  "https://getcomposer.org/doc/",
		},
		"go.mod": {
			name: "Go",
			url:  "https://golang.org/doc/",
		},
		"requirements.txt": {
			name: "Python",
			url:  "https://docs.python.org/3/",
		},
		"package.json": {
			name: "Node.js",
			url:  "https://nodejs.org/docs/latest/api/",
		},
		"angular.json": {
			name: "Angular",
			url:  "https://angular.io/docs",
		},
		"vue.config.js": {
			name: "Vue",
			url:  "https://vuejs.org/guide/introduction.html",
		},
		"tailwind.config.js": {
			name: "Tailwind CSS",
			url:  "https://tailwindcss.com/docs",
		},
		"nuxt.config.js": {
			name: "Nuxt.js",
			url:  "https://nuxtjs.org/docs/",
		},
		"next.config.js": {
			name: "Next.js",
			url:  "https://nextjs.org/docs/",
		},
		"svelte.config.js": {
			name: "Svelte",
			url:  "https://svelte.dev/docs",
		},
		"webpack.config.js": {
			name: "Webpack",
			url:  "https://webpack.js.org/concepts/",
		},
		"babel.config.js": {
			name: "Babel",
			url:  "https://babeljs.io/docs/",
		},
		"jest.config.js": {
			name: "Jest",
			url:  "https://jestjs.io/docs/",
		},
		"cypress.json": {
			name: "Cypress",
			url:  "https://docs.cypress.io/",
		},
		"tsconfig.json": {
			name: "TypeScript",
			url:  "https://www.typescriptlang.org/docs/",
		},
		".eslintrc.js": {
			name: "ESLint",
			url:  "https://eslint.org/docs/user-guide/",
		},
		".prettierrc": {
			name: "Prettier",
			url:  "https://prettier.io/docs/en/",
		},
		"Gemfile": {
			name: "Ruby Bundler",
			url:  "https://bundler.io/guides/",
		},
		"Pipfile": {
			name: "Pipenv",
			url:  "https://pipenv.pypa.io/en/latest/",
		},
		"poetry.lock": {
			name: "Poetry",
			url:  "https://python-poetry.org/docs/",
		},
		"Cargo.toml": {
			name: "Rust",
			url:  "https://doc.rust-lang.org/book/",
		},
		"mix.exs": {
			name: "Elixir",
			url:  "https://elixir-lang.org/docs.html",
		},
		"stack.yaml": {
			name: "Haskell",
			url:  "https://www.haskell.org/documentation/",
		},
	}

	// Check for existence of config files
	for fileName, info := range configFiles {
		// Look for the config file anywhere in the project
		var found bool
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.EqualFold(info.Name(), fileName) {
				found = true
				return filepath.SkipAll
			}
			return nil
		})

		if found {
			depsMap[info.name] = DependencyDocs{
				Name:   info.name,
				DocURL: info.url,
			}
		}
	}

	// Check for package.json dependencies
	if _, err := os.Stat("package.json"); err == nil {
		content, err := ioutil.ReadFile("package.json")
		if err == nil {
			var packageJSON map[string]interface{}
			if err := json.Unmarshal(content, &packageJSON); err == nil {
				// Define common npm packages and their docs
				npmPackages := map[string]struct {
					name string
					url  string
				}{
					"express": {
						name: "Express",
						url:  "https://expressjs.com/en/4x/api.html",
					},
					"react": {
						name: "React",
						url:  "https://react.dev/reference/react",
					},
					"vue": {
						name: "Vue",
						url:  "https://vuejs.org/guide/introduction.html",
					},
					"svelte": {
						name: "Svelte",
						url:  "https://svelte.dev/docs",
					},
					"@angular/core": {
						name: "Angular",
						url:  "https://angular.io/docs",
					},
					"tailwindcss": {
						name: "Tailwind CSS",
						url:  "https://tailwindcss.com/docs",
					},
					"bootstrap": {
						name: "Bootstrap",
						url:  "https://getbootstrap.com/docs/",
					},
					"jquery": {
						name: "jQuery",
						url:  "https://api.jquery.com/",
					},
					"next": {
						name: "Next.js",
						url:  "https://nextjs.org/docs/",
					},
					"nuxt": {
						name: "Nuxt.js",
						url:  "https://nuxtjs.org/docs/",
					},
					"redux": {
						name: "Redux",
						url:  "https://redux.js.org/introduction/getting-started",
					},
					"mobx": {
						name: "MobX",
						url:  "https://mobx.js.org/README.html",
					},
					"axios": {
						name: "Axios",
						url:  "https://axios-http.com/docs/intro",
					},
					"lodash": {
						name: "Lodash",
						url:  "https://lodash.com/docs/",
					},
					"moment": {
						name: "Moment.js",
						url:  "https://momentjs.com/docs/",
					},
					"d3": {
						name: "D3.js",
						url:  "https://d3js.org/",
					},
					"three": {
						name: "Three.js",
						url:  "https://threejs.org/docs/",
					},
					"socket.io": {
						name: "Socket.IO",
						url:  "https://socket.io/docs/",
					},
					"mongoose": {
						name: "Mongoose",
						url:  "https://mongoosejs.com/docs/",
					},
					"typeorm": {
						name: "TypeORM",
						url:  "https://typeorm.io/",
					},
					"sequelize": {
						name: "Sequelize",
						url:  "https://sequelize.org/",
					},
					"prisma": {
						name: "Prisma",
						url:  "https://www.prisma.io/docs/",
					},
					"storybook": {
						name: "Storybook",
						url:  "https://storybook.js.org/docs/",
					},
					"jest": {
						name: "Jest",
						url:  "https://jestjs.io/docs/",
					},
					"mocha": {
						name: "Mocha",
						url:  "https://mochajs.org/",
					},
					"chai": {
						name: "Chai",
						url:  "https://www.chaijs.com/",
					},
					"cypress": {
						name: "Cypress",
						url:  "https://docs.cypress.io/",
					},
					"playwright": {
						name: "Playwright",
						url:  "https://playwright.dev/docs/intro",
					},
					"webpack": {
						name: "Webpack",
						url:  "https://webpack.js.org/concepts/",
					},
					"babel": {
						name: "Babel",
						url:  "https://babeljs.io/docs/",
					},
					"eslint": {
						name: "ESLint",
						url:  "https://eslint.org/docs/user-guide/",
					},
					"prettier": {
						name: "Prettier",
						url:  "https://prettier.io/docs/en/",
					},
					"sass": {
						name: "Sass",
						url:  "https://sass-lang.com/documentation",
					},
					"less": {
						name: "Less",
						url:  "https://lesscss.org/",
					},
					"styled-components": {
						name: "styled-components",
						url:  "https://styled-components.com/docs",
					},
					"emotion": {
						name: "Emotion",
						url:  "https://emotion.sh/docs/introduction",
					},
					"material-ui": {
						name: "Material-UI",
						url:  "https://mui.com/material-ui/getting-started/",
					},
					"@mui/material": {
						name: "Material-UI",
						url:  "https://mui.com/material-ui/getting-started/",
					},
					"antd": {
						name: "Ant Design",
						url:  "https://ant.design/docs/react/introduce",
					},
				}

				// Check dependencies and devDependencies
				for section := range map[string]string{"dependencies": "prod", "devDependencies": "dev"} {
					if deps, ok := packageJSON[section].(map[string]interface{}); ok {
						for pkgName := range deps {
							if info, exists := npmPackages[pkgName]; exists {
								depsMap[info.name] = DependencyDocs{
									Name:   info.name,
									DocURL: info.url,
								}
							}
						}
					}
				}
			}
		}
	}

	// Check for composer.json dependencies
	if _, err := os.Stat("composer.json"); err == nil {
		content, err := ioutil.ReadFile("composer.json")
		if err == nil {
			var data map[string]interface{}
			if err := json.Unmarshal(content, &data); err == nil {
				// Define common PHP packages and their docs
				phpPackages := map[string]struct {
					name string
					url  string
				}{
					"laravel/framework": {
						name: "Laravel",
						url:  "https://laravel.com/docs",
					},
					"symfony/symfony": {
						name: "Symfony",
						url:  "https://symfony.com/doc/current/",
					},
					"slim/slim": {
						name: "Slim Framework",
						url:  "https://www.slimframework.com/docs/",
					},
					"cakephp/cakephp": {
						name: "CakePHP",
						url:  "https://book.cakephp.org/",
					},
					"codeigniter/framework": {
						name: "CodeIgniter",
						url:  "https://codeigniter.com/user_guide/",
					},
					"yiisoft/yii2": {
						name: "Yii Framework",
						url:  "https://www.yiiframework.com/doc/guide/",
					},
					"laminas/laminas-mvc": {
						name: "Laminas Framework",
						url:  "https://docs.laminas.dev/",
					},
					"zendframework/zend-mvc": {
						name: "Zend Framework",
						url:  "https://docs.laminas.dev/",
					},
					"doctrine/orm": {
						name: "Doctrine ORM",
						url:  "https://www.doctrine-project.org/projects/doctrine-orm/en/current/index.html",
					},
					"illuminate/database": {
						name: "Laravel Eloquent",
						url:  "https://laravel.com/docs/eloquent",
					},
					"twig/twig": {
						name: "Twig",
						url:  "https://twig.symfony.com/doc/",
					},
					"smarty/smarty": {
						name: "Smarty",
						url:  "https://www.smarty.net/docs/en/",
					},
					"phpunit/phpunit": {
						name: "PHPUnit",
						url:  "https://phpunit.de/documentation.html",
					},
					"squizlabs/php_codesniffer": {
						name: "PHP_CodeSniffer",
						url:  "https://github.com/squizlabs/PHP_CodeSniffer/wiki",
					},
					"phpstan/phpstan": {
						name: "PHPStan",
						url:  "https://phpstan.org/user-guide/getting-started",
					},
					"nunomaduro/larastan": {
						name: "Larastan",
						url:  "https://github.com/nunomaduro/larastan",
					},
					"inertiajs/inertia-laravel": {
						name: "InertiaJS",
						url:  "https://inertiajs.com/",
					},
					"ishanvyas22/cakephp-inertiajs": {
						name: "InertiaJS",
						url:  "https://inertiajs.com/",
					},
					"inertiajs/inertia": {
						name: "InertiaJS",
						url:  "https://inertiajs.com/",
					},
					"guzzlehttp/guzzle": {
						name: "Guzzle",
						url:  "https://docs.guzzlephp.org/",
					},
					"monolog/monolog": {
						name: "Monolog",
						url:  "https://github.com/Seldaek/monolog/blob/main/doc/01-usage.md",
					},
					"league/flysystem": {
						name: "Flysystem",
						url:  "https://flysystem.thephpleague.com/docs/",
					},
					"firebase/php-jwt": {
						name: "PHP-JWT",
						url:  "https://github.com/firebase/php-jwt",
					},
					"erusev/parsedown": {
						name: "Parsedown",
						url:  "https://github.com/erusev/parsedown",
					},
					"spatie/laravel-permission": {
						name: "Laravel Permission",
						url:  "https://spatie.be/docs/laravel-permission/",
					},
				}

				// Check require section
				if req, ok := data["require"].(map[string]interface{}); ok {
					for pkgName := range req {
						if info, exists := phpPackages[strings.ToLower(pkgName)]; exists {
							depsMap[info.name] = DependencyDocs{
								Name:   info.name,
								DocURL: info.url,
							}
						}
					}
				}

				// Add Composer to dependencies
				depsMap["Composer"] = DependencyDocs{
					Name:   "Composer",
					DocURL: "https://getcomposer.org/doc/",
				}
			}
		}
	}

	// Check for requirements.txt dependencies
	if _, err := os.Stat("requirements.txt"); err == nil {
		content, err := ioutil.ReadFile("requirements.txt")
		if err == nil {
			reqContent := string(content)
			lines := strings.Split(reqContent, "\n")

			// Define common Python packages and their docs
			pythonPackages := map[string]struct {
				name string
				url  string
			}{
				"flask": {
					name: "Flask",
					url:  "https://flask.palletsprojects.com/",
				},
				"django": {
					name: "Django",
					url:  "https://docs.djangoproject.com/",
				},
				"fastapi": {
					name: "FastAPI",
					url:  "https://fastapi.tiangolo.com/",
				},
				"tornado": {
					name: "Tornado",
					url:  "https://www.tornadoweb.org/en/stable/",
				},
				"pyramid": {
					name: "Pyramid",
					url:  "https://docs.pylonsproject.org/projects/pyramid/",
				},
				"sanic": {
					name: "Sanic",
					url:  "https://sanic.dev/",
				},
				"sqlalchemy": {
					name: "SQLAlchemy",
					url:  "https://docs.sqlalchemy.org/",
				},
				"django-rest-framework": {
					name: "Django REST Framework",
					url:  "https://www.django-rest-framework.org/",
				},
				"djangorestframework": {
					name: "Django REST Framework",
					url:  "https://www.django-rest-framework.org/",
				},
				"pandas": {
					name: "pandas",
					url:  "https://pandas.pydata.org/docs/",
				},
				"numpy": {
					name: "NumPy",
					url:  "https://numpy.org/doc/",
				},
				"scipy": {
					name: "SciPy",
					url:  "https://docs.scipy.org/doc/scipy/",
				},
				"matplotlib": {
					name: "Matplotlib",
					url:  "https://matplotlib.org/stable/contents.html",
				},
				"scikit-learn": {
					name: "scikit-learn",
					url:  "https://scikit-learn.org/stable/user_guide.html",
				},
				"tensorflow": {
					name: "TensorFlow",
					url:  "https://www.tensorflow.org/api_docs",
				},
				"pytorch": {
					name: "PyTorch",
					url:  "https://pytorch.org/docs/stable/index.html",
				},
				"torch": {
					name: "PyTorch",
					url:  "https://pytorch.org/docs/stable/index.html",
				},
				"keras": {
					name: "Keras",
					url:  "https://keras.io/api/",
				},
				"requests": {
					name: "Requests",
					url:  "https://docs.python-requests.org/",
				},
				"beautifulsoup4": {
					name: "Beautiful Soup",
					url:  "https://www.crummy.com/software/BeautifulSoup/bs4/doc/",
				},
				"scrapy": {
					name: "Scrapy",
					url:  "https://docs.scrapy.org/",
				},
				"pytest": {
					name: "pytest",
					url:  "https://docs.pytest.org/",
				},
				"celery": {
					name: "Celery",
					url:  "https://docs.celeryq.dev/",
				},
				"pillow": {
					name: "Pillow",
					url:  "https://pillow.readthedocs.io/",
				},
				"opencv-python": {
					name: "OpenCV",
					url:  "https://docs.opencv.org/4.x/d6/d00/tutorial_py_root.html",
				},
			}

			// Parse each line to extract package name
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				// Handle different requirement formats
				parts := strings.Split(line, "==")
				if len(parts) < 1 {
					continue
				}

				packageName := strings.ToLower(strings.TrimSpace(parts[0]))
				if info, exists := pythonPackages[packageName]; exists {
					depsMap[info.name] = DependencyDocs{
						Name:   info.name,
						DocURL: info.url,
					}
				}
			}

			// Add Python to dependencies
			depsMap["Python"] = DependencyDocs{
				Name:   "Python",
				DocURL: "https://docs.python.org/3/",
			}
		}
	}

	// Check for go.mod
	if _, err := os.Stat("go.mod"); err == nil {
		// For demonstration, we add a dependency for Go documentation.
		depsMap["Go"] = DependencyDocs{
			Name:   "Go",
			DocURL: "https://golang.org/doc/",
		}

		// Check the go.mod for Go dependencies
		content, err := ioutil.ReadFile("go.mod")
		if err == nil {
			goModContent := string(content)

			// Map of common Go packages to check for
			goPackages := map[string]struct {
				name string
				url  string
			}{
				"github.com/gofiber/fiber": {
					name: "Fiber",
					url:  "https://docs.gofiber.io/",
				},
				"github.com/gin-gonic/gin": {
					name: "Gin",
					url:  "https://gin-gonic.com/docs/",
				},
				"github.com/gorilla/mux": {
					name: "Gorilla Mux",
					url:  "https://pkg.go.dev/github.com/gorilla/mux",
				},
				"github.com/labstack/echo": {
					name: "Echo",
					url:  "https://echo.labstack.com/guide/",
				},
				"gorm.io/gorm": {
					name: "GORM",
					url:  "https://gorm.io/docs/",
				},
				"github.com/jinzhu/gorm": {
					name: "GORM",
					url:  "https://gorm.io/docs/",
				},
			}

			// Check for each Go package
			for pkg, info := range goPackages {
				if strings.Contains(goModContent, pkg) {
					depsMap[info.name] = DependencyDocs{
						Name:   info.name,
						DocURL: info.url,
					}
				}
			}
		}
	}

	// Check for main.py and other Python-specific files
	hasPythonMain := false
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/directories we can't access
		}
		if !info.IsDir() && info.Name() == "main.py" {
			hasPythonMain = true
			return filepath.SkipAll // Stop searching once we find it
		}
		return nil
	})

	if hasPythonMain {
		depsMap["Python"] = DependencyDocs{
			Name:   "Python",
			DocURL: "https://docs.python.org/3/",
		}
	}

	// Additional check for Bootstrap - recursively search for bootstrap CSS files
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files/directories we can't access
		}
		if !info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "bootstrap") && strings.HasSuffix(strings.ToLower(info.Name()), ".css") {
			depsMap["Bootstrap"] = DependencyDocs{
				Name:   "Bootstrap",
				DocURL: "https://getbootstrap.com/docs/",
			}
			return filepath.SkipAll // Stop searching once we find one
		}
		return nil
	})

	// Convert map to slice
	var deps []DependencyDocs
	for _, dep := range depsMap {
		deps = append(deps, dep)
	}

	if len(deps) == 0 {
		return nil, fmt.Errorf("no known dependencies found")
	}

	return deps, nil
}
